# tradinghour — Claude guide

Go and Rust libraries for global market trading hours, sharing the same embedded YAML calendars. Given a unix timestamp + market enum, tells you whether the market is open and which session (premarket / regular / postmarket / overnight). Also returns full day timelines.

## Status

Implemented. The original design is at [docs/superpowers/specs/2026-04-17-tradinghour-design.md](docs/superpowers/specs/2026-04-17-tradinghour-design.md); current source and [Rust API and maintenance](docs/rust-library.md) describe later changes. Read those before changing code or data.

## MVP markets

NASDAQ (including BOAT overnight), HKEX, China A-Share (SSE+SZSE), TSE (Tokyo), KRX (Korea), Rates (interest rate products), Metals (spot gold/silver).

## Layout (target)

- `tradinghour.go`, `market.go`, `loader.go`, `phase.go`, `holiday.go` — flat package, one purpose per file.
- `Cargo.toml`, `build.rs`, `rust/src/` — native Rust API, calendar evaluation, and embedding of the same data tree.
- `testdata/parity.json`, `internal/parity/`, `rust/examples/parity.rs` — shared input and independent Go/Rust public-API output for `scripts/check_parity.py`.
- `data/markets/*.yaml` — weekly schedule + half-day schedule per market.
- `data/holidays/<market>/<year>.yaml` — annual holiday calendars (`type: closed` or `type: half_day`).
- `scripts/refresh_holidays.py` + `.github/workflows/refresh-holidays.yml` — yearly PR to refresh holidays from `exchange_calendars`. Never auto-merged.

## Conventions

- All schedule times are in the market's local tz, written as `"HH:MM"` in YAML. Overnight phases use a `"HH:MM+1"` suffix for "next day".
- Public API returns `time.Time` values anchored in the market's `*time.Location`. Callers convert to UTC if they want.
- Data is embedded via `go:embed` and parsed once at `init()`. No runtime file I/O.
- Rust embeds the same data with `include_str!`, parses once, and retains local timezones. Preserve Go/Rust parity for every valid shipped calendar; run both suites and `scripts/check_parity.py` after changes.
- Holiday data is reviewed by a human before merge — do not bypass PR review, even for "obvious" updates.

## Adding a market (post-MVP)

1. Drop a `data/markets/<market>.yaml`.
2. Drop `data/holidays/<market>/<year>.yaml` for each year.
3. Add the `MarketType` constant.
4. Add the `exchange_calendars` mapping in `scripts/refresh_holidays.py`.
5. Add table-driven tests covering boundaries, weekend, holiday, half-day.
6. Add the Rust `Market`/`ALL` entry and shared parity input, then run both language suites, parity, and `cargo package --locked`.
