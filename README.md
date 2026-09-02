# tradinghour

Go library for global market trading hours. Tells you whether a market is open at a given instant, which session is active (premarket / regular / postmarket / overnight), and the full trading timeline for any date.

## Install

```bash
go get github.com/uranuswch/trading-hour
```

## Usage

```go
import (
    "fmt"
    "time"

    th "github.com/uranuswch/trading-hour"
)

func main() {
    status, _ := th.IsOpen(time.Now().Unix(), th.MarketNASDAQ)
    fmt.Printf("NASDAQ open=%v session=%s\n", status.Open, status.Session)

    ds, _ := th.Timeline(time.Now(), th.MarketHKEX)
    for _, p := range ds.Phases {
        fmt.Printf("  %s  %s -> %s\n", p.Session, p.Start, p.End)
    }
}
```

## Supported markets

| Market          | Constant                 | Timezone            |
|-----------------|--------------------------|---------------------|
| NASDAQ + NYSE   | `th.MarketNASDAQ`        | America/New_York    |
| HKEX (equity)   | `th.MarketHKEX`          | Asia/Hong_Kong      |
| SSE + SZSE      | `th.MarketChinaAShare`   | Asia/Shanghai       |
| Tokyo (TSE)     | `th.MarketTSE`           | Asia/Tokyo          |
| Taiwan (TWSE)   | `th.MarketTWSE`          | Asia/Taipei         |
| Korea (KRX)     | `th.MarketKRX`           | Asia/Seoul          |
| Forex           | `th.MarketFX`            | America/New_York    |
| CME             | `th.MarketCME`           | America/New_York    |
| ICE             | `th.MarketICE`           | America/New_York    |
| FXCM UK Oil     | `th.MarketFXCMUKOil`     | UTC                 |
| FXCM US Oil     | `th.MarketFXCMUSOil`     | UTC                 |
| Rates           | `th.MarketRates`         | America/New_York    |
| Metals          | `th.MarketMetals`        | America/New_York    |

NASDAQ includes the Blue Ocean ATS overnight session (8pm–4am ET, Sun–Thu).

TWSE covers board-lot equity trading Monday–Friday: `regular` 09:00–13:30
(no lunch break) and `postmarket` 14:00–14:30 in `Asia/Taipei` (UTC+8).
The postmarket phase represents the after-hours fixed-price order window;
orders are matched once at 14:30. Phase starts are inclusive and ends exclusive.
Pre-open order collection, odd-lot and block trading are outside this schedule.
The embedded 2026 calendar includes all 18 scheduled weekday closures, including
the February 12–13 settlement-only days, and has no half-days.
Sources: [TWSE trading mechanism](https://www.twse.com.tw/en/products/system/trading.html)
and [2026 holiday calendar](https://www.twse.com.tw/holidaySchedule/holidaySchedule?response=html&queryYear=2026).

## Web Dashboard

A live market-status dashboard is included in `web/static/` and served by a small Go HTTP server in `cmd/server/`.

**Run:**

```bash
go run ./cmd/server/
# → listening on http://localhost:8080
```

Set `PORT` to override the default port:

```bash
PORT=9000 go run ./cmd/server/
```

The dashboard auto-refreshes every 30 seconds and shows:

- **Pills** — open/partial/closed status for all 13 markets at a glance
- **Spotlight** — countdown to the next market open
- **Side drawer** — 24-hour timeline bar, session list, and date picker for any market

The server binary embeds `web/static/` via `go:embed`, so it has no working-directory dependency and can be deployed as a single self-contained binary.

## Data

Market schedules and holiday calendars live in `data/` as YAML and are embedded into the binary via `go:embed`. A GitHub Action runs yearly (November 15) to open a PR generating next-year holidays from [`exchange_calendars`](https://pypi.org/project/exchange-calendars/). PRs require human review before merge.

## Design

See [docs/superpowers/specs/2026-04-17-tradinghour-design.md](docs/superpowers/specs/2026-04-17-tradinghour-design.md).

## License

MIT — see [LICENSE](LICENSE).
