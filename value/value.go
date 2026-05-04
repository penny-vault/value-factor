// Copyright 2021-2026
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package value

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/penny-vault/pvbt/asset"
	"github.com/penny-vault/pvbt/data"
	"github.com/penny-vault/pvbt/engine"
	"github.com/penny-vault/pvbt/portfolio"
)

//go:embed README.md
var description string

// ValueFactor implements an earnings yield (EBIT/EV) value factor strategy.
// It ranks stocks by EV/EBIT ratio and buys the cheapest stocks (lowest ratio).
type ValueFactor struct {
	IndexName   string `pvbt:"index" desc:"Stock index universe to select from" default:"us-tradable" suggest:"SPX=SPX|NDX=NDX|us-tradable=us-tradable"`
	TopHoldings int    `pvbt:"top-holdings" desc:"Number of cheapest stocks to hold" default:"50" suggest:"SPX=50|NDX=10"`
	SectorCap   int    `pvbt:"sector-cap" desc:"Maximum holdings per GICS sector (0 disables the cap)" default:"4" suggest:"Vanilla=0"`
	MinFScore   int    `pvbt:"min-fscore" desc:"Minimum Piotroski F-score (0-9) required to hold a stock; 0 disables the screen" default:"6" suggest:"Vanilla=0"`
}

func (s *ValueFactor) Name() string {
	return "Value Factor"
}

func (s *ValueFactor) Setup(_ *engine.Engine) {}

func (s *ValueFactor) Describe() engine.StrategyDescription {
	return engine.StrategyDescription{
		ShortCode:   "val",
		Description: description,
		Source:      "https://doi.org/10.1111/j.1540-6261.1992.tb04398.x",
		Version:     "1.1.0",
		VersionDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
		Schedule:    "@quarterend",
		Benchmark:   "VFINX",
	}
}

func (s *ValueFactor) Compute(ctx context.Context, eng *engine.Engine, strategyPortfolio portfolio.Portfolio, batch *portfolio.Batch) error {
	// 1. Resolve the index universe to per-stock asset records (sector-populated).
	currentDate := eng.CurrentDate()

	members := eng.IndexUniverse(s.IndexName).Assets(currentDate)
	if len(members) == 0 {
		return nil
	}

	// 2. Fetch EV/EBIT ratio for all members at the current date.
	valuationDF, err := eng.FetchAt(ctx, members, currentDate, []data.Metric{data.EVtoEBIT})
	if err != nil {
		return fmt.Errorf("failed to fetch EV/EBIT ratios: %w", err)
	}

	// 3. Rank stocks by EV/EBIT ascending (lowest = cheapest = best value).
	//    Filter out stocks with non-positive EV/EBIT (unprofitable or missing data).
	candidates := make([]stockValuation, 0, len(members))
	for _, stock := range members {
		ratio := valuationDF.Value(stock, data.EVtoEBIT)

		// Skip stocks with missing, negative, or zero EV/EBIT (unprofitable or bad data).
		if math.IsNaN(ratio) || ratio <= 0 {
			continue
		}

		candidates = append(candidates, stockValuation{stock: stock, evEBIT: ratio})
	}

	// Optional Piotroski F-score quality screen.
	if s.MinFScore > 0 {
		candidates, err = s.applyFScoreFilter(ctx, eng, candidates, currentDate)
		if err != nil {
			return fmt.Errorf("piotroski filter: %w", err)
		}
	}

	// Sort ascending by EV/EBIT (cheapest first).
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].evEBIT < candidates[j].evEBIT
	})

	topCount := min(s.TopHoldings, len(candidates))
	if topCount == 0 {
		return nil
	}

	// 4. Greedy selection. With SectorCap > 0, fill in EV/EBIT order while
	// enforcing a per-GICS-sector quota; otherwise take the cheapest topCount.
	selected := make([]stockValuation, 0, topCount)
	sectorCount := map[asset.Sector]int{}

	for _, sv := range candidates {
		if len(selected) == topCount {
			break
		}

		if s.SectorCap > 0 && sectorCount[sv.stock.Sector] >= s.SectorCap {
			continue
		}

		selected = append(selected, sv)
		sectorCount[sv.stock.Sector]++
	}

	if len(selected) == 0 {
		return nil
	}

	// 5. Equal weight across selected stocks.
	weight := 1.0 / float64(len(selected))
	targets := make(map[asset.Asset]float64, len(selected))

	for _, sv := range selected {
		targets[sv.stock] = weight
	}

	var justification string
	if s.SectorCap > 0 {
		justification = fmt.Sprintf("top %d/%d cheapest by EV/EBIT from %s, max %d per sector",
			len(selected), len(candidates), s.IndexName, s.SectorCap)
	} else {
		justification = fmt.Sprintf("top %d/%d cheapest by EV/EBIT from %s",
			len(selected), len(candidates), s.IndexName)
	}

	batch.Annotate("universe-size", fmt.Sprintf("%d", len(candidates)))
	batch.Annotate("justification", justification)

	allocation := portfolio.Allocation{
		Date:          currentDate,
		Members:       targets,
		Justification: justification,
	}

	if err := batch.RebalanceTo(ctx, allocation); err != nil {
		return fmt.Errorf("rebalance failed: %w", err)
	}

	return nil
}

// stockValuation pairs a stock with its EV/EBIT ratio for ranking.
type stockValuation struct {
	stock  asset.Asset
	evEBIT float64
}

// piotroskiMetrics is the bundle of fundamentals needed to compute the
// nine-component Piotroski F-score plus the year-ago comparison points.
var piotroskiMetrics = []data.Metric{
	data.NetIncome,
	data.ROA,
	data.NetCashFlowFromOperations,
	data.DebtNonCurrent,
	data.CurrentRatio,
	data.SharesBasic,
	data.GrossMargin,
	data.AssetTurnover,
}

// applyFScoreFilter drops candidates whose Piotroski F-score is below
// MinFScore. The score uses the most recent quarter-end whose 10-Q is
// reliably filed (1-quarter lag from currentDate) and the same quarter
// one year prior for the YoY comparisons.
func (s *ValueFactor) applyFScoreFilter(
	ctx context.Context,
	eng *engine.Engine,
	candidates []stockValuation,
	currentDate time.Time,
) ([]stockValuation, error) {
	if len(candidates) == 0 {
		return candidates, nil
	}

	currKey := priorQuarterEnd(currentDate)
	prevKey := currKey.AddDate(-1, 0, 0)

	assets := make([]asset.Asset, len(candidates))
	for i, c := range candidates {
		assets[i] = c.stock
	}

	currDF, err := eng.FetchFundamentalsByDateKey(ctx, assets, piotroskiMetrics, currKey,
		engine.WithAsOfDate(currentDate))
	if err != nil {
		return nil, fmt.Errorf("fetch fundamentals at %s: %w", currKey.Format("2006-01-02"), err)
	}

	prevDF, err := eng.FetchFundamentalsByDateKey(ctx, assets, piotroskiMetrics, prevKey,
		engine.WithAsOfDate(currentDate))
	if err != nil {
		return nil, fmt.Errorf("fetch fundamentals at %s: %w", prevKey.Format("2006-01-02"), err)
	}

	filtered := make([]stockValuation, 0, len(candidates))
	for _, c := range candidates {
		if piotroskiScore(currDF, prevDF, c.stock) >= s.MinFScore {
			filtered = append(filtered, c)
		}
	}

	return filtered, nil
}

// piotroskiScore returns the 0-9 Piotroski F-score for a single stock.
// Missing or NaN inputs fail the corresponding signal (conservative).
func piotroskiScore(curr, prev *data.DataFrame, stock asset.Asset) int {
	cNI := curr.Value(stock, data.NetIncome)
	cROA := curr.Value(stock, data.ROA)
	cCFO := curr.Value(stock, data.NetCashFlowFromOperations)
	cDebt := curr.Value(stock, data.DebtNonCurrent)
	cCR := curr.Value(stock, data.CurrentRatio)
	cShares := curr.Value(stock, data.SharesBasic)
	cGM := curr.Value(stock, data.GrossMargin)
	cAT := curr.Value(stock, data.AssetTurnover)

	pDebt := prev.Value(stock, data.DebtNonCurrent)
	pCR := prev.Value(stock, data.CurrentRatio)
	pShares := prev.Value(stock, data.SharesBasic)
	pGM := prev.Value(stock, data.GrossMargin)
	pAT := prev.Value(stock, data.AssetTurnover)

	score := 0
	if !math.IsNaN(cNI) && cNI > 0 {
		score++
	}

	if !math.IsNaN(cROA) && cROA > 0 {
		score++
	}

	if !math.IsNaN(cCFO) && cCFO > 0 {
		score++
	}

	if !math.IsNaN(cCFO) && !math.IsNaN(cNI) && cCFO > cNI {
		score++
	}

	if !math.IsNaN(cDebt) && !math.IsNaN(pDebt) && cDebt < pDebt {
		score++
	}

	if !math.IsNaN(cCR) && !math.IsNaN(pCR) && cCR > pCR {
		score++
	}

	if !math.IsNaN(cShares) && !math.IsNaN(pShares) && cShares <= pShares {
		score++
	}

	if !math.IsNaN(cGM) && !math.IsNaN(pGM) && cGM > pGM {
		score++
	}

	if !math.IsNaN(cAT) && !math.IsNaN(pAT) && cAT > pAT {
		score++
	}

	return score
}

// priorQuarterEnd returns the calendar quarter-end at least one quarter
// before currentDate, used as the fundamentals date_key. For a quarter-end
// date this returns the previous quarter-end (e.g. 2024-03-28 -> 2023-12-31).
func priorQuarterEnd(currentDate time.Time) time.Time {
	quarter := (int(currentDate.Month()) - 1) / 3

	year := currentDate.Year()
	if quarter == 0 {
		year--
		quarter = 4
	}

	monthEnd := []time.Month{time.March, time.June, time.September, time.December}
	dayEnd := []int{31, 30, 30, 31}

	return time.Date(year, monthEnd[quarter-1], dayEnd[quarter-1], 0, 0, 0, 0, currentDate.Location())
}
