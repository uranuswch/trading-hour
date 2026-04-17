# tradinghour — Claude guide

Go library for global market trading hours. Given a unix timestamp + market enum, tells you whether the market is open and which session (premarket / regular / postmarket / overnight). Also returns full day timelines.

## Status

Pre-implementation. The approved design is at [docs/superpowers/specs/2026-04-17-tradinghour-design.md](docs/superpowers/specs/2026-04-17-tradinghour-design.md). Read that first before changing any code or data.

## MVP markets

NASDAQ (including BOAT overnight), HKEX, China A-Share (SSE+SZSE), TSE (Tokyo), KRX (Korea), Rates (interest rate products), Metals (spot gold/silver).

## Layout (target)

- `tradinghour.go`, `market.go`, `loader.go`, `phase.go`, `holiday.go` — flat package, one purpose per file.
- `data/markets/*.yaml` — weekly schedule + half-day schedule per market.
- `data/holidays/<market>/<year>.yaml` — annual holiday calendars (`type: closed` or `type: half_day`).
- `scripts/refresh_holidays.py` + `.github/workflows/refresh-holidays.yml` — yearly PR to refresh holidays from `exchange_calendars`. Never auto-merged.

## Conventions

- All schedule times are in the market's local tz, written as `"HH:MM"` in YAML. Overnight phases use a `"HH:MM+1"` suffix for "next day".
- Public API returns `time.Time` values anchored in the market's `*time.Location`. Callers convert to UTC if they want.
- Data is embedded via `go:embed` and parsed once at `init()`. No runtime file I/O.
- Holiday data is reviewed by a human before merge — do not bypass PR review, even for "obvious" updates.

## Adding a market (post-MVP)

1. Drop a `data/markets/<market>.yaml`.
2. Drop `data/holidays/<market>/<year>.yaml` for each year.
3. Add the `MarketType` constant.
4. Add the `exchange_calendars` mapping in `scripts/refresh_holidays.py`.
5. Add table-driven tests covering boundaries, weekend, holiday, half-day.
