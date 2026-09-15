# Native Rust library

The `trading-hour` Cargo package and the existing Go module share `data/markets`
and `data/holidays`. `build.rs` inventories those files deterministically and
generates `include_str!` entries. The registry is parsed once through `OnceLock`;
queries do not read files, call Go, or contact a service.

## API

| Go | Rust | Contract |
| --- | --- | --- |
| `IsOpen(unixSec, market)` | `is_open(unix_sec, market)` | Absolute Unix seconds; checks today and previous-day spillover |
| `Timeline(date, market)` | `timeline(NaiveDate, market)` | Civil date in the market timezone; phases start on that date |
| `NextOpen(unixSec, market)` | `next_open(unix_sec, market)` | Next strictly future phase start within 15 calendar days |
| `NextClose(unixSec, market)` | `next_close(unix_sec, market)` | End of current phase, or next phase when closed |
| `MarketLocation(market)` | `market_location(market)` | Exchange-local timezone |
| Embedded holiday files | `holiday_years(market)` | Sorted years with an embedded file, even if its holiday list is empty |

Returned `DateTime<chrono_tz::Tz>` values retain the market timezone. `Market`
supports exact Go identifiers through `FromStr` and `Display`. `Session` includes
closed, premarket, regular, postmarket, overnight and continuous. Invalid market
identifiers, calendar data, out-of-range timestamps, and exhausted searches return
typed errors. A non-unique/nonexistent local boundary in embedded data returns
`InvalidDateTime`; the shipped schedules do not use such DST transition times.

Starts are inclusive and ends exclusive. `next_open` also returns adjacent session
starts while a market is already open, preserving the existing Go API. Timeline
dates are local civil dates, so callers should explicitly convert an instant into
the market timezone before taking its date.

## Shared semantics

- Weekly schedules use local wall-clock time, including `+1` overnight ends.
- The latest applicable `effective_from` replaces the complete weekly schedule;
  omitted weekdays are closed. Holiday and half-day rules retain precedence.
- NASDAQ overnight belongs to the following trading day. A holiday evening may
  open ahead of a normal day; the evening before a closed holiday is suppressed.
- A holiday flag describes the calendar date, so its timeline may still contain
  a next-trading-day overnight phase.
- Outside the embedded holiday years, the weekly schedule applies, as in Go.
  `holiday_years` makes this limitation observable; it does not invent holidays.

Calendar policy does not decide per-security eligibility, broker permissions,
halts, stale quotes, or whether a source supports a particular market session.
Those checks remain with the consumer. Consumers that must stop publishing at
close should query the calendar again immediately before selecting cached prices.

## Maintenance and validation

Keep all calendar edits in the existing `data/` tree and review them through the
existing holiday workflow. No copied YAML belongs in a consuming repository.
To add a market, follow the Go checklist, add a Rust `Market` variant and `ALL`
entry, and extend `testdata/parity.json`. To add a year, add the reviewed holiday
files and extend the parity range. The build automatically includes added files.

`scripts/check_parity.py` uses the shared input file to compare both public APIs
across every day, phase boundary minus/at/plus one second, and fixed UTC probes.
Rust also has independent expected-value tests adapted from the existing Go
market tests. `cargo package` verifies the packaged archive can rebuild with its
embedded calendars, without access to the source checkout.

Consumers should pin a reviewed Git revision. Calendar changes require dependency
upgrades and rebuilt binaries in both languages. Publishing to crates.io is a
separate release operation; Git consumers do not require it.
