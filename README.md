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
| Korea (KRX)     | `th.MarketKRX`           | Asia/Seoul          |
| Forex           | `th.MarketFX`            | America/New_York    |
| CME             | `th.MarketCME`           | America/New_York    |
| ICE             | `th.MarketICE`           | America/New_York    |
| FXCM UK Oil     | `th.MarketFXCMUKOil`     | UTC                 |
| FXCM US Oil     | `th.MarketFXCMUSOil`     | UTC                 |
| Rates           | `th.MarketRates`         | America/New_York    |
| Metals          | `th.MarketMetals`        | America/New_York    |

NASDAQ includes the Blue Ocean ATS overnight session (8pm–4am ET, Sun–Thu).

## Data

Market schedules and holiday calendars live in `data/` as YAML and are embedded into the binary via `go:embed`. A GitHub Action runs yearly (November 15) to open a PR generating next-year holidays from [`exchange_calendars`](https://pypi.org/project/exchange-calendars/). PRs require human review before merge.

## Design

See [docs/superpowers/specs/2026-04-17-tradinghour-design.md](docs/superpowers/specs/2026-04-17-tradinghour-design.md).

## License

MIT — see [LICENSE](LICENSE).
