# `tradinghour` — Go global market trading hours library

**Date:** 2026-04-17
**Status:** Approved, pending implementation plan

## 1. Purpose

A Go library that answers two questions for any supported equity market:

1. At a given instant (unix timestamp), is the market open, and which session is active?
2. For a given date, what is the full trading timeline — including pre/post-market, overnight, lunch breaks, half-days, and holidays?

The library is inspired by [tradinghours.com](https://docs.tradinghours.com/3.x/enterprise/trading-hours) but is self-contained: all schedule and holiday data ships embedded in the Go binary, with a GitHub Action that opens a yearly PR to refresh holidays.

## 2. MVP scope

### 2.1 Supported markets

| MarketType        | Exchange(s)                  | Timezone           |
|-------------------|------------------------------|--------------------|
| `NASDAQ`          | Nasdaq + NYSE + Blue Ocean ATS | America/New_York  |
| `HKEX`            | Hong Kong Exchange (equity)  | Asia/Hong_Kong     |
| `ChinaAShare`     | SSE + SZSE                   | Asia/Shanghai      |
| `TSE`             | Tokyo Stock Exchange         | Asia/Tokyo         |
| `KRX`             | Korea Exchange               | Asia/Seoul         |

### 2.2 Session model per market

| Market        | premarket       | regular                                   | postmarket      | overnight         |
|---------------|-----------------|-------------------------------------------|-----------------|-------------------|
| NASDAQ        | 04:00–09:30     | 09:30–16:00                               | 16:00–20:00     | 20:00→04:00 next day, Sun–Thu (BOAT) |
| HKEX          | —               | 09:30–12:00, 13:00–16:10 (lunch 12:00–13:00) | —            | —                 |
| ChinaAShare   | —               | 09:30–11:30, 13:00–15:00 (lunch 11:30–13:00) | —            | —                 |
| TSE           | —               | 09:00–11:30, 12:30–15:30 (lunch 11:30–12:30) | —            | —                 |
| KRX           | 08:00–09:00     | 09:00–15:30                               | 15:40–18:00    | —                 |

Pre-open auctions (HKEX 09:00–09:30, ChinaAShare 09:15–09:25, KRX call auction) are rolled into the adjacent `regular` phase for MVP simplicity. Post-MVP can split them out as their own session type if needed.

### 2.3 Non-goals for MVP

- Futures, options, ETFs with venue-specific rules.
- Non-equity venues (crypto, FX, commodities).
- Predicting holidays — only looking them up from shipped data.
- Sub-session microstructure beyond the table above (separate auction vs. continuous phases).
- Historical schedule changes (e.g. TSE 15:00→15:30 cutover on 2024-11-05). MVP ships one schedule per market; version stamps can be added post-MVP.

## 3. Public API

```go
package tradinghour

import "time"

type MarketType string

const (
    MarketNASDAQ      MarketType = "NASDAQ"
    MarketHKEX        MarketType = "HKEX"
    MarketChinaAShare MarketType = "ChinaAShare"
    MarketTSE         MarketType = "TSE"
    MarketKRX         MarketType = "KRX"
)

type Session string

const (
    SessionClosed     Session = "closed"
    SessionPreMarket  Session = "premarket"
    SessionRegular    Session = "regular"
    SessionPostMarket Session = "postmarket"
    SessionOvernight  Session = "overnight"
)

type Status struct {
    Open    bool
    Session Session
    Market  MarketType
}

type Phase struct {
    Session Session
    Start   time.Time // in the market's local timezone
    End     time.Time // may be on a later calendar day than Start (overnight)
}

type DaySchedule struct {
    Date        time.Time   // midnight in market-local tz
    Market      MarketType
    Phases      []Phase     // sorted by Start; may contain a phase whose End is on the next day
    IsHoliday   bool        // full closure
    IsHalfDay   bool
    HolidayName string      // empty if not a holiday
}

// Core
func IsOpen(unixSec int64, m MarketType) (Status, error)
func Timeline(date time.Time, m MarketType) (DaySchedule, error)

// Convenience
func NextOpen(unixSec int64, m MarketType) (time.Time, error)
func NextClose(unixSec int64, m MarketType) (time.Time, error)

// Errors
var ErrUnknownMarket = errors.New("tradinghour: unknown market")
```

### 3.1 Semantics

- `IsOpen(unixSec, m)` — unix seconds is absolute. The implementation converts to the market's local time and checks phases for today *and* yesterday (overnight spillover).
- `Timeline(date, m)` — only Y/M/D of `date` is read, interpreted in the market's local timezone. `hour/min/sec/loc` of the input are ignored.
- `Timeline` returns phases that *start* on the given date. An overnight phase starting 20:00 Mon and ending 04:00 Tue appears in Monday's timeline, not Tuesday's. Consequence: `IsOpen(Tue 02:00, NASDAQ)` returns `overnight` even though `Timeline(Tue)` does not include that phase. Callers who need "what's live at time T" should use `IsOpen`; `Timeline` answers "what trading happens during day D".
- `NextOpen(unixSec, m)` — nearest future instant at which `IsOpen` transitions to true. Walks forward day-by-day up to 15 calendar days (covers Golden Week + weekends).
- `NextClose(unixSec, m)` — if market is open, the end of the current phase (e.g. for NASDAQ at 15:59 returns 16:00, the end of the regular session — not 20:00 end-of-postmarket). If closed, the end of the next open phase.
- Errors: unknown `MarketType` returns `ErrUnknownMarket`. No other error paths in MVP (data load failures would panic at package init, which is the right signal since the library is unusable).

## 4. Data files

All data is shipped in the repo under `data/` and embedded via `go:embed`. Parsed once at `init()` into in-memory structs.

### 4.1 Market schedule — `data/markets/<market>.yaml`

```yaml
market: NASDAQ
timezone: America/New_York
weekly_schedule:
  monday:
    - {session: premarket,  start: "04:00", end: "09:30"}
    - {session: regular,    start: "09:30", end: "16:00"}
    - {session: postmarket, start: "16:00", end: "20:00"}
    - {session: overnight,  start: "20:00", end: "04:00+1"}
  tuesday:   # same as monday
    - {session: premarket,  start: "04:00", end: "09:30"}
    - {session: regular,    start: "09:30", end: "16:00"}
    - {session: postmarket, start: "16:00", end: "20:00"}
    - {session: overnight,  start: "20:00", end: "04:00+1"}
  # ... wednesday, thursday likewise
  friday:    # BOAT does not start Friday
    - {session: premarket,  start: "04:00", end: "09:30"}
    - {session: regular,    start: "09:30", end: "16:00"}
    - {session: postmarket, start: "16:00", end: "20:00"}
  saturday: []
  sunday:
    - {session: overnight,  start: "20:00", end: "04:00+1"}
half_day_schedule:
  - {session: regular,    start: "09:30", end: "13:00"}
  - {session: postmarket, start: "13:00", end: "17:00"}
```

**Time format:** `"HH:MM"` (24h). Suffix `+1` means "next calendar day" — used for overnight phases.

**`half_day_schedule`** is the replacement phase list used when a date is flagged as `half_day` in the holiday file. Omitted for markets without half-days.

### 4.2 Holiday calendar — `data/holidays/<market>/<year>.yaml`

```yaml
market: NASDAQ
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",          type: closed}
  - {date: "2026-01-19", name: "MLK Day",                 type: closed}
  - {date: "2026-11-27", name: "Day after Thanksgiving",  type: half_day}
  - {date: "2026-12-24", name: "Christmas Eve",           type: half_day}
  - {date: "2026-12-25", name: "Christmas Day",           type: closed}
```

**`type`** is `closed` (full closure, empty phases) or `half_day` (use market's `half_day_schedule`).

## 5. Timezone semantics

- Every computation is anchored in the market's local time via `time.LoadLocation(<tz>)`.
- DST transitions are handled by Go's `time` package — no special logic needed.
- `Phase.Start` / `Phase.End` are always in the market's `*time.Location`. Callers can `.UTC()` them freely.
- Overnight phase end times are constructed as `time.Date(y, m, d+1, endH, endM, 0, 0, loc)`, which correctly handles month/year wrap and DST boundaries.

## 6. Core algorithms

### 6.1 `IsOpen(unixSec, m)`

```
1. t = time.Unix(unixSec, 0).In(market.tz)
2. Look up today's phases = materialize(market, t.Date())
3. Look up yesterday's phases = materialize(market, t.Date()-1day)
4. For each phase in (today ∪ yesterday) where phase.Start <= t < phase.End:
     return Status{Open: true, Session: phase.Session, Market: m}
5. Return Status{Open: false, Session: SessionClosed, Market: m}
```

### 6.2 `Timeline(date, m)`

```
1. Normalize date to market-local midnight: d = time.Date(y, mo, dy, 0,0,0,0, market.tz)
2. phases, isHoliday, isHalfDay, name = materialize(market, d)
3. Return DaySchedule{d, m, phases, isHoliday, isHalfDay, name}
```

### 6.3 `materialize(market, date)`

```
1. weekday = date.Weekday()
2. base = market.weekly_schedule[weekday]
3. holiday = market.holidayLookup[date]
4. if holiday == closed:   return [], true, false, holiday.name
5. if holiday == half_day: return applyHalfDay(market), false, true, holiday.name
6. return instantiatePhases(base, date), false, false, ""
```

`instantiatePhases` turns `{"04:00", "09:30"}` into real `time.Time` values on `date` in `market.tz`, handling `+1` suffix by advancing the day component.

### 6.4 `NextOpen` / `NextClose`

Linear scan forward, starting from today, up to 15 days. Each day is materialized once. First phase with `Start > t` (NextOpen) or `End > t` while currently open (NextClose) wins.

## 7. Package layout

```
/tradinghour.go           // public types, IsOpen, Timeline, NextOpen, NextClose, ErrUnknownMarket
/market.go                // Market struct: weekly schedule + holidays; materialize()
/loader.go                // go:embed FS, YAML parsing, package-init wiring
/phase.go                 // Phase model; time materialization incl. "+1" handling
/holiday.go               // Holiday types + lookup map
/tradinghour_test.go      // Public API table-driven tests
/market_test.go
/phase_test.go
/data/markets/nasdaq.yaml
/data/markets/hkex.yaml
/data/markets/china-ashare.yaml
/data/markets/tse.yaml
/data/markets/krx.yaml
/data/holidays/nasdaq/2026.yaml
/data/holidays/hkex/2026.yaml
/data/holidays/china-ashare/2026.yaml
/data/holidays/tse/2026.yaml
/data/holidays/krx/2026.yaml
/scripts/refresh_holidays.py
/.github/workflows/refresh-holidays.yml
/README.md
/go.mod
```

Each Go file has one clear purpose: loader knows nothing about the public API; API files call into `Market.materialize()`.

## 8. GitHub Action: holiday refresh

**File:** `.github/workflows/refresh-holidays.yml`

**Triggers:**
- Scheduled: `cron: '0 0 15 11 *'` (November 15th annually).
- Manual: `workflow_dispatch`.

**Job:**
1. Checkout repo.
2. Set up Python 3.11.
3. `pip install exchange_calendars pyyaml`.
4. Run `scripts/refresh_holidays.py --year=$(date +%Y -d '+1 year')`.
5. Use `peter-evans/create-pull-request` to open a PR with:
   - Title: `chore: refresh holidays for <year>`
   - Body: links to official calendars (NYSE, HKEX, SSE, JPX, KRX) + checklist of half-days flagged for manual review.
   - Labels: `data`, `needs-review`.

**Script:** `scripts/refresh_holidays.py` maps each `MarketType` → `exchange_calendars` name, generates a YAML file per market for the target year, flags half-days from the library's early-close metadata. No auto-merge — every PR requires human review before shipping.

## 9. Testing strategy

- **Unit tests (table-driven) per market covering:**
  - Regular hours boundaries (open/closed at start, middle, end).
  - Pre/post-market transitions (NASDAQ, KRX).
  - Lunch break (HKEX, ChinaAShare, TSE).
  - Overnight crossing midnight (NASDAQ BOAT on Mon-Tue boundary).
  - Weekends (Fri evening, Sat, Sun — with NASDAQ Sunday overnight).
  - Full-closure holidays (Christmas NASDAQ, Lunar New Year ChinaAShare — multi-day).
  - Half-days (NASDAQ Nov 27 2026 early close, HKEX Lunar NY eve).
  - DST transitions (NASDAQ spring-forward + fall-back).
- **Golden-file tests** for `Timeline` output on known tricky dates.
- **Round-trip** `IsOpen` consistency with `Timeline`: for every phase in a day's timeline, `IsOpen` at Start should return open, at End-1ns open, at End closed.
- **No network** in tests; everything embedded.

## 10. Open questions / post-MVP

- Auction sub-sessions (pre-open, closing auction) as first-class phases.
- Historical schedule versioning (handle TSE 15:30 cutover, NYSE early-2020 COVID disruptions).
- Per-security override (e.g. ETFs with different NYSE hours).
- Extending to futures venues (CME, SGX, ICE).
- Subscribing to a paid data source (TradingHours.com) for authoritative updates.
