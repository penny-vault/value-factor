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
	IndexName   string `pvbt:"index" desc:"Stock index universe to select from" default:"SPX" suggest:"SPX=SPX|NDX=NDX"`
	TopHoldings int    `pvbt:"top-holdings" desc:"Number of cheapest stocks to hold" default:"50" suggest:"SPX=50|NDX=10"`
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
		Version:     "1.0.0",
		VersionDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		Schedule:    "@quarterend",
		Benchmark:   "VFINX",
	}
}

func (s *ValueFactor) Compute(ctx context.Context, eng *engine.Engine, strategyPortfolio portfolio.Portfolio, batch *portfolio.Batch) error {
	// 1. Get the index universe for the current date.
	indexUniverse := eng.IndexUniverse(s.IndexName)

	// 2. Fetch EV/EBIT ratio for all members at the current date.
	valuationDF, err := indexUniverse.At(ctx, data.EVtoEBIT)
	if err != nil {
		return fmt.Errorf("failed to fetch EV/EBIT ratios: %w", err)
	}

	if valuationDF.Len() == 0 {
		return nil
	}

	// 3. Rank stocks by EV/EBIT ascending (lowest = cheapest = best value).
	//    Filter out stocks with non-positive EV/EBIT (unprofitable or missing data).
	type stockValuation struct {
		stock  asset.Asset
		evEBIT float64
	}

	var candidates []stockValuation

	for _, stock := range valuationDF.AssetList() {
		ratio := valuationDF.Value(stock, data.EVtoEBIT)

		// Skip stocks with missing, negative, or zero EV/EBIT (unprofitable or bad data).
		if math.IsNaN(ratio) || ratio <= 0 {
			continue
		}

		candidates = append(candidates, stockValuation{stock: stock, evEBIT: ratio})
	}

	// Sort ascending by EV/EBIT (cheapest first).
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].evEBIT < candidates[j].evEBIT
	})

	topCount := s.TopHoldings
	if topCount > len(candidates) {
		topCount = len(candidates)
	}

	if topCount == 0 {
		return nil
	}

	selected := candidates[:topCount]

	// 4. Equal weight across selected stocks.
	weight := 1.0 / float64(topCount)
	members := make(map[asset.Asset]float64, topCount)

	for _, sv := range selected {
		members[sv.stock] = weight
	}

	justification := fmt.Sprintf("top %d/%d cheapest by EV/EBIT from %s", topCount, len(candidates), s.IndexName)

	batch.Annotate("universe-size", fmt.Sprintf("%d", len(candidates)))
	batch.Annotate("justification", justification)

	allocation := portfolio.Allocation{
		Date:          eng.CurrentDate(),
		Members:       members,
		Justification: justification,
	}

	if err := batch.RebalanceTo(ctx, allocation); err != nil {
		return fmt.Errorf("rebalance failed: %w", err)
	}

	return nil
}
