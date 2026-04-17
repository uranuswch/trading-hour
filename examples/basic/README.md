# Trading Hour Examples

This directory contains example code demonstrating how to use the `tradinghour` library.

## Examples

- **basic/** - Simple examples showing basic API usage
  - `is_open.go` - Check if markets are open at a given time
  - `timeline.go` - Get full day schedule for a market
  - `new_markets.go` - Examples for the new Rates and Metals markets

## Running Examples

```bash
# Run an example
go run examples/basic/is_open.go

# Run all examples
for f in examples/basic/*.go; do go run "$f"; done
```

## API Overview

### Check Market Status

```go
import (
    "time"
    "github.com/yourusername/tradinghour"
)

// Check if NASDAQ is open now
status, err := tradinghour.IsOpen(time.Now().Unix(), tradinghour.MarketNASDAQ)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("NASDAQ open: %v, session: %s\n", status.Open, status.Session)
```

### Get Daily Timeline

```go
// Get full schedule for a specific date
date := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
schedule, err := tradinghour.Timeline(date, tradinghour.MarketNASDAQ)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Schedule: %+v\n", schedule)
```

## Supported Markets

- **NASDAQ** - US equities (including BOAT overnight)
- **HKEX** - Hong Kong Stock Exchange
- **ChinaAShare** - Chinese A-shares (SSE + SZSE)
- **TSE** - Tokyo Stock Exchange
- **KRX** - Korea Exchange
- **FX** - Spot foreign exchange (24/5)
- **CME** - CME futures (24/5 with daily maintenance)
- **ICE** - ICE futures (24/5 with daily maintenance)
- **FXCMUKOil** - FXCM UKOIL CFD
- **FXCMUSOil** - FXCM USOIL CFD
- **Rates** - Interest rate products (business hours)
- **Metals** - Spot gold/silver (24/5 with daily maintenance)
