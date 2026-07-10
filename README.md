# Fasal Nivesh (फसल निवेश) — "Crop Investment"

Farmers lose real money every season not because they grow the wrong
crop, but because they sell it at the wrong time, in the wrong place,
without knowing any better. Fasal Nivesh closes that information gap.

## What it is

Fasal Nivesh watches real market and weather data continuously and turns
it into three plain answers for a farmer:

- **Sell now, or wait?** — a short-term price forecast for their crop
- **Which market is actually worth the trip?** — nearby markets ranked
  by what they'll really put in the farmer's pocket after transport cost,
  not just the sticker price
- **What should I plant next season?** — an estimate based on how
  weather and prices have historically moved together

## Why it exists

A farmer selling produce today usually knows only the price at the one
market in front of them, on the one day they show up. They don't know
that a market 30km away is paying more this week, or that a forecasted
dry spell means prices are about to rise if they can hold their harvest
a few more days. That single gap in information — not a lack of good
produce — is what quietly erodes a season's income. All of the data
needed to close that gap is already public and free; it's just never
been turned into a straight answer a farmer can act on.

## How it solves the problem

The system never stops watching. Every day, it pulls fresh market prices
and weather data, learns from the patterns building up over time, and
turns that into forecasts and recommendations — delivered in a form that
works even for someone without a smartphone or a data plan.

```mermaid
flowchart TD
    A["Mandi prices<br/><small>Daily market data</small>"] --> C["Continuous collection<br/><small>Always up to date</small>"]
    B["Weather patterns<br/><small>Rain and temperature</small>"] --> C
    C --> D["Prediction engine<br/><small>Learns from patterns</small>"]
    D --> E["Price forecast<br/><small>Sell now or wait</small>"]
    D --> F["Best market<br/><small>Highest real profit</small>"]
    D --> G["Planting advice<br/><small>What to grow next</small>"]
    E --> H["Reaches every farmer<br/><small>Web and SMS alerts</small>"]
    F --> H
    G --> H
```

The core idea is a continuous loop, not a one-time report: today's data
sharpens tomorrow's forecast, and every recommendation is delivered in
the way that reaches a farmer, not just the way that's easiest to build.

## License

MIT — this is a learning/portfolio project built entirely on public,
open government data.
