# tradinghour

Go and Rust libraries for global market trading hours. Both embed the same reviewed YAML calendars and tell you whether a market is open at a given instant, which session is active (premarket / regular / postmarket / overnight), and the full trading timeline for any date.

## Install

### Go

```bash
go get github.com/uranuswch/trading-hour
```

### Rust

The native Rust crate lives in this repository. Pin a reviewed full commit SHA:

```toml
[dependencies]
trading-hour = { git = "https://github.com/uranuswch/trading-hour", rev = "<reviewed-commit-sha>" }
```

```rust
use trading_hour::{is_open, Market, Session};

fn main() -> Result<(), trading_hour::Error> {
    let status = is_open(1789455600, Market::Krx)?;
    assert!(status.open);
    assert_eq!(status.session, Session::PostMarket);
    Ok(())
}
```

Rust 1.85 or newer is required. See [Rust API and maintenance](docs/rust-library.md)
for timeline queries, calendar coverage, packaging, and cross-language verification.

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

KRX uses `Asia/Seoul` (UTC+9): `premarket` 08:00–09:00,
`regular` 09:00–15:30, and `postmarket` 15:40–20:00 from September 14, 2026.
Earlier dates retain the 18:00 postmarket close. The postmarket phase combines
closing-price trading (15:40–16:00) with continuous after-market trading
(16:00–20:00), which replaces the former 16:00–18:00 single-price auctions.
Phase starts are inclusive and ends exclusive; weekends and holidays remain closed.
The continuous after-market covers eligible equities and initially excludes
ETFs, ETNs, and other restricted securities. This is a market-level schedule;
callers must check individual security eligibility separately.
Sources: [KRX trading hours](https://global.krx.co.kr/contents/GLB/06/0602/0602020204/GLB0602020204T1.jsp)
and [September 14 launch notice](https://www.samsungpop.com/ux/kor/customer/notice/notice/noticeViewContent.do?MenuSeqNo=24420).

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

Schedules may include `weekly_schedule_overrides`, each with an `effective_from`
date (`YYYY-MM-DD`, inclusive in the market's timezone) and a complete
`weekly_schedule`. The latest applicable override replaces the base weekly
schedule; omitted weekdays are closed. Earlier dates use the base schedule.
Holiday closures and `half_day_schedule` still apply. See `data/markets/krx.yaml`
for an example. Other markets retain their existing unversioned schedules.

Go and Rust use this directory directly; there is no second calendar copy. Updating
these files takes effect when consumers upgrade their pinned dependency and rebuild.
Calendar coverage is limited to the embedded years. Outside those years both
implementations use the weekly schedule; Rust exposes `holiday_years` so consumers
can report missing coverage. A market-level open phase does not establish an
individual instrument's eligibility, entitlement, halt status, or quote freshness.

## Verification

```sh
go test -race ./...
cargo fmt --all -- --check
cargo clippy --locked --all-targets --all-features -- -D warnings
cargo test --locked --all-targets --all-features
cargo test --locked --doc
python3 scripts/check_parity.py
cargo package --locked
```

The parity check compares all 13 markets over the entire embedded 2026 year,
including surrounding year boundaries, every phase start/end and fixed UTC probes.
It verifies status, timeline metadata and offsets, next open, and next close against
the existing Go implementation. All query tests run offline with embedded data.

## Design

See [docs/superpowers/specs/2026-04-17-tradinghour-design.md](docs/superpowers/specs/2026-04-17-tradinghour-design.md).

## License

MIT — see [LICENSE](LICENSE).
