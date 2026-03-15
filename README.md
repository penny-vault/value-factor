# Value Factor (Earnings Yield)

The **Value Factor** strategy is based on the foundational work of Fama and French (1992) on value investing and Joel Greenblatt's "Magic Formula" (2005) which uses earnings yield (EBIT/EV) as the primary value signal. The strategy ranks stocks by their EV/EBIT ratio and buys the cheapest stocks -- those with the lowest enterprise value relative to their earnings.

## Rules

1. On the last trading day of each quarter, fetch the EV/EBIT ratio for each stock in the universe.
2. Filter out stocks with negative or zero EBIT (unprofitable companies).
3. Rank all remaining stocks by EV/EBIT ascending (lowest ratio = cheapest = highest earnings yield).
4. Select the top holdings (cheapest stocks).
5. Equal weight across selected stocks.
6. Hold all positions until the next quarterly rebalance.

Quarterly rebalancing aligns with earnings release cycles and reduces turnover compared to monthly rebalancing.

## Parameters

- **Index**: Which stock universe to draw from (default: S&P 500)
- **Top Holdings**: Number of cheapest stocks to hold (default: 50)

## References

- Fama, E. and French, K. (1992). "The Cross-Section of Expected Stock Returns." *Journal of Finance*, 47(2), 427-465.
- Greenblatt, J. (2005). *The Little Book That Beats the Market*. John Wiley & Sons.
