# Value Factor (Earnings Yield)

The **Value Factor** strategy is based on the foundational work of Fama and French (1992) on value investing and Joel Greenblatt's "Magic Formula" (2005) which uses earnings yield (EBIT/EV) as the primary value signal. The strategy ranks stocks by their EV/EBIT ratio and buys the cheapest stocks -- those with the lowest enterprise value relative to their earnings.

## Rules

1. On the last trading day of each quarter, fetch the EV/EBIT ratio for each stock in the universe.
2. Filter out stocks with negative or zero EBIT (unprofitable companies).
3. If the Piotroski F-score screen is enabled, drop names whose 9-component F-score (computed from the most recent 10-Q with a one-quarter filing buffer, against the same quarter one year prior) is below the threshold.
4. Rank all remaining stocks by EV/EBIT ascending (lowest ratio = cheapest = highest earnings yield).
5. Walk the ranked list from cheapest down. If the sector cap is enabled, skip a stock once its GICS sector has reached the cap; otherwise take the cheapest names until the holdings count is met.
6. Equal weight across selected stocks.
7. Hold all positions until the next quarterly rebalance.

Quarterly rebalancing aligns with earnings release cycles and reduces turnover compared to monthly rebalancing.

The sector cap forces sector diversification (raw EV/EBIT ranks tend to cluster in cyclicals such as energy and financials). The Piotroski F-score is a binary quality screen designed by Joseph Piotroski (2000) to separate value stocks with improving fundamentals from "value traps" with deteriorating ones.

## Parameters

- **Index**: Which stock universe to draw from (default: us-tradable)
- **Top Holdings**: Number of cheapest stocks to hold (default: 50)
- **Sector Cap**: Maximum holdings per GICS sector (default: 4; set to 0 to disable)
- **Min F-Score**: Minimum Piotroski F-score required to hold a stock, 0-9 (default: 6; set to 0 to disable)

## Presets

- **Vanilla** turns off both the sector cap and the F-score screen, reproducing the unscreened Greenblatt earnings-yield strategy. Useful as an academic baseline.

## References

- Fama, E. and French, K. (1992). "The Cross-Section of Expected Stock Returns." *Journal of Finance*, 47(2), 427-465.
- Greenblatt, J. (2005). *The Little Book That Beats the Market*. John Wiley & Sons.
- Piotroski, J. (2000). "Value Investing: The Use of Historical Financial Statement Information to Separate Winners from Losers." *Journal of Accounting Research*, 38, 1-41.
