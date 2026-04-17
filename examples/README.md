# Trading Hour Examples

This directory contains example code demonstrating how to use the `tradinghour` library.

## Quick Start

```go
import (
    "time"
    "github.com/uranuswch/tradinghour"
)

// Check if NASDAQ is open right now
status, err := tradinghour.IsOpen(time.Now().Unix(), tradinghour.MarketNASDAQ)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("NASDAQ is open: %v\n", status.Open)
```

## Examples

### Basic Usage (`basic/`)

| Example | Description |
|---------|-------------|
| `is_open.go` | Check if markets are open at a given time |
| `timeline.go` | Get full day schedule for a market |
| `new_markets.go` | Examples for Rates and Metals markets |

### Running Examples

```bash
# Run a single example
go run examples/basic/is_open.go

# Run all basic examples
cd examples/basic
for f in *.go; do go run "$f"; echo; done
```

## Supported Markets

### Equity Markets
- **NASDAQ** - US equities (premarket, regular, postmarket, BOAT overnight)
- **HKEX** - Hong Kong Stock Exchange
- **ChinaAShare** - Chinese A-shares (SSE + SZSE)
- **TSE** - Tokyo Stock Exchange
- **KRX** - Korea Exchange

### FX & Commodities
- **FX** - Spot foreign exchange (24/5 continuous)
- **CME** - CME futures (24/5 with daily maintenance 17:00-18:00 ET)
- **ICE** - ICE futures (24/5 with daily maintenance 18:00-20:00 ET)
- **FXCMUKOil** - FXCM UKOIL CFD (Mon-Fri 01:00-22:00 UTC)
- **FXCMUSOil** - FXCM USOIL CFD (Sun 23:00 - Fri 21:45 UTC)

### New Markets
- **Rates** - Interest rate products (Mon-Fri 08:00-17:00 ET, business hours)
- **Metals** - Spot gold/silver (24/5 with daily maintenance 17:00-18:00 ET)

## API Reference

### `IsOpen(unixSec int64, m MarketType) (Status, error)`

Check if a market is open at a given Unix timestamp.

**Returns:**
- `Status.Open` - true if market is open
- `Status.Session` - current session (premarket, regular, postmarket, overnight, continuous, closed)
- `Status.Market` - the market queried

**Example:**
```go
status, err := tradinghour.IsOpen(time.Now().Unix(), tradinghour.MarketNASDAQ)
if status.Open {
    fmt.Println("Market is open in", status.Session, "session")
}
```

### `Timeline(date time.Time, m MarketType) (DaySchedule, error)`

Get the full trading schedule for a specific date.

**Returns:**
- `DaySchedule.Date` - the date (market's local timezone)
- `DaySchedule.Phases` - all trading phases for the day
- `DaySchedule.IsHoliday` - true if this is a holiday
- `DaySchedule.IsHalfDay` - true if this is a half-day
- `DaySchedule.HolidayName` - name of the holiday (if applicable)

**Example:**
```go
date := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
schedule, err := tradinghour.Timeline(date, tradinghour.MarketNASDAQ)
for _, phase := range schedule.Phases {
    fmt.Printf("%s: %s - %s\n", phase.Session, phase.Start, phase.End)
}
```

## Session Types

| Session | Description |
|---------|-------------|
| `premarket` | Pre-market trading |
| `regular` | Regular trading hours |
| `postmarket` | Post-market trading |
| `overnight` | Overnight session (spans to next day) |
| `continuous` | Continuous trading (24/5 markets) |
| `closed` | Market is closed |

## Timezone Handling

All timestamps are Unix seconds (UTC). The library handles timezone conversion internally based on the market's configured timezone:

- US markets (NASDAQ, CME, Metals, Rates) → `America/New_York`
- Asian markets (HKEX, ChinaAShare, TSE, KRX) → Local timezone
- FXCM markets → `UTC`

You can pass any Unix timestamp; the library will convert it to the market's local timezone for evaluation.

**Example:**
```go
// Create a time in New York, pass as Unix timestamp
ny, _ := time.LoadLocation("America/New_York")
localTime := time.Date(2026, 1, 5, 10, 0, 0, 0, ny)
status, _ := tradinghour.IsOpen(localTime.Unix(), tradinghour.MarketNASDAQ)
// The library correctly interprets this as 10am ET
```

## More Examples

### Check Multiple Markets

```go
markets := []tradinghour.MarketType{
    tradinghour.MarketNASDAQ,
    tradinghour.MarketFX,
    tradinghour.MarketMetals,
}

for _, market := range markets {
    status, _ := tradinghour.IsOpen(time.Now().Unix(), market)
    fmt.Printf("%s: %v\n", market, status.Open)
}
```

### Find Next Open Time

```go
// Check every 5 minutes until market opens
for {
    status, _ := tradinghour.IsOpen(time.Now().Unix(), tradinghour.MarketNASDAQ)
    if status.Open {
        fmt.Println("NASDAQ is now open!")
        break
    }
    time.Sleep(5 * time.Minute)
}
```

### Holiday Handling

```go
// Check Christmas 2026
christmas := time.Date(2026, 12, 25, 10, 0, 0, 0, time.UTC)
schedule, _ := tradinghour.Timeline(christmas, tradinghour.MarketNASDAQ)

if schedule.IsHoliday {
    fmt.Printf("Market closed for %s\n", schedule.HolidayName)
}
```

## Tips

1. **Use `Timeline` for debugging** - See the full day schedule to understand market hours
2. **Check holidays** - Always verify `IsHoliday` before assuming regular hours apply
3. **Handle timezones** - Remember Unix timestamps are UTC; the library handles timezone conversion
4. **Session types vary** - Equity markets use premarket/regular/postmarket; FX/Metals use continuous
5. **Weekend closures** - Most markets are closed on weekends (except 24/5 markets like FX)

## Contributing

Found an issue or have a suggestion? Please open an issue or pull request on GitHub.
