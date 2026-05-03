package value_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/penny-vault/pvbt/asset"
	"github.com/penny-vault/pvbt/data"
	"github.com/penny-vault/pvbt/engine"
	"github.com/penny-vault/pvbt/portfolio"
	"github.com/penny-vault/value-factor/value"
)

var _ = Describe("ValueFactor", func() {
	var (
		ctx       context.Context
		snap      *data.SnapshotProvider
		nyc       *time.Location
		startDate time.Time
		endDate   time.Time
	)

	BeforeEach(func() {
		ctx = context.Background()

		var err error
		nyc, err = time.LoadLocation("America/New_York")
		Expect(err).NotTo(HaveOccurred())

		snap, err = data.NewSnapshotProvider("testdata/snapshot.db")
		Expect(err).NotTo(HaveOccurred())

		startDate = time.Date(2023, 10, 1, 0, 0, 0, 0, nyc)
		endDate = time.Date(2025, 1, 1, 0, 0, 0, 0, nyc)
	})

	AfterEach(func() {
		if snap != nil {
			snap.Close()
		}
	})

	runBacktest := func() portfolio.Portfolio {
		strategy := &value.ValueFactor{IndexName: "SPX"}
		acct := portfolio.New(
			portfolio.WithCash(100000, startDate),
			portfolio.WithAllMetrics(),
		)

		eng := engine.New(strategy,
			engine.WithDataProvider(snap),
			engine.WithAssetProvider(snap),
			engine.WithAccount(acct),
		)

		result, err := eng.Backtest(ctx, startDate, endDate)
		Expect(err).NotTo(HaveOccurred())
		return result
	}

	It("produces expected returns and risk metrics", func() {
		result := runBacktest()

		summary, err := result.Summary()
		Expect(err).NotTo(HaveOccurred())
		Expect(summary.TWRR).To(BeNumerically("~", 0.1466, 0.01))
		Expect(summary.MaxDrawdown).To(BeNumerically(">", -0.15), "max drawdown should be better than -15%")

		Expect(result.Value()).To(BeNumerically("~", 114661, 500))
	})

	It("rebalances on all quarter-end dates", func() {
		result := runBacktest()
		txns := result.Transactions()

		rebalanceDates := map[string]bool{}
		for _, t := range txns {
			if t.Type == asset.BuyTransaction || t.Type == asset.SellTransaction {
				rebalanceDates[t.Date.In(nyc).Format("2006-01-02")] = true
			}
		}

		Expect(rebalanceDates).To(HaveKey("2023-12-29")) // Q4 2023 end
		Expect(rebalanceDates).To(HaveKey("2024-03-28")) // Q1 2024 end
		Expect(rebalanceDates).To(HaveKey("2024-06-28")) // Q2 2024 end
		Expect(rebalanceDates).To(HaveKey("2024-09-30")) // Q3 2024 end
		Expect(rebalanceDates).To(HaveKey("2024-12-31")) // Q4 2024 end
	})

	It("buys approximately 50 stocks on the initial rebalance", func() {
		result := runBacktest()
		txns := result.Transactions()

		// Count unique tickers bought on the first rebalance (Q4 2023 end)
		firstRebalanceBuys := map[string]bool{}
		for _, t := range txns {
			if t.Type == asset.BuyTransaction {
				d := t.Date.In(nyc).Format("2006-01-02")
				if d == "2023-12-29" {
					firstRebalanceBuys[t.Asset.Ticker] = true
				}
			}
		}

		Expect(len(firstRebalanceBuys)).To(BeNumerically(">=", 45),
			"should buy at least 45 stocks on first rebalance (got %d)", len(firstRebalanceBuys))
		Expect(len(firstRebalanceBuys)).To(BeNumerically("<=", 51),
			"should buy at most 51 stocks on first rebalance (got %d)", len(firstRebalanceBuys))
	})

	It("trades a meaningful number of unique stocks", func() {
		result := runBacktest()
		txns := result.Transactions()

		tickers := map[string]bool{}
		for _, t := range txns {
			if t.Type == asset.BuyTransaction {
				tickers[t.Asset.Ticker] = true
			}
		}

		Expect(len(tickers)).To(BeNumerically(">=", 60),
			"should trade at least 60 unique stocks across all rebalances")
	})
})
