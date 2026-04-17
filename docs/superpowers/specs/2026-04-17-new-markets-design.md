# New Markets Design: FX, CME, ICE, FXCM, Rates, Metals

## 1. Goal

Extend the `tradinghour` library with seven new `MarketType` constants covering spot FX, CME futures, ICE futures, FXCM OTC CFDs (UK oil + US oil), interest rates, and spot metals. All data is derived from the [Pyth Network market hours reference](https://docs.pyth.network/price-feeds/core/market-hours).

## 2. New Session Type

Add `SessionContinuous` to `tradinghour.go`:

```go
SessionContinuous Session = "continuous"
```

Used for open windows in 24/5 markets. `Rates` continues to use `SessionRegular` (discrete business-hours session). `SessionContinuous` is purely a string value — no changes required in `loader.go`, `market.go`, `phase.go`, or `holiday.go`.

## 3. New MarketType Constants

```go
MarketFX          MarketType = "FX"
MarketCME         MarketType = "CME"
MarketICE         MarketType = "ICE"
MarketFXCMUKOil   MarketType = "FXCMUKOil"
MarketFXCMUSOil   MarketType = "FXCMUSOil"
MarketRates       MarketType = "Rates"
MarketMetals      MarketType = "Metals"
```

## 4. Market Schedule Details

### 4.1 FX — Spot Foreign Exchange

- **Timezone:** `America/New_York`
- **Source:** Pyth "FX" section
- **Pattern:** True 24/5 — no maintenance gaps. Sunday open at 17:00 ET; Friday close at 17:00 ET.
- **Holidays:** Christmas Day, New Year's Day (hand-curated; FX observes almost no holidays)

```yaml
market: FX
timezone: America/New_York
weekly_schedule:
  sunday:    [{session: continuous, start: "17:00", end: "00:00+1"}]
  monday:    &fx_full [{session: continuous, start: "00:00", end: "00:00+1"}]
  tuesday:   *fx_full
  wednesday: *fx_full
  thursday:  *fx_full
  friday:    [{session: continuous, start: "00:00", end: "17:00"}]
  saturday:  []
```

### 4.2 CME — WTI Crude Oil Futures

- **Timezone:** `America/New_York`
- **Source:** Pyth "WTI Crude (CME)" section
- **Pattern:** 24/5 with 1-hour daily maintenance window Mon–Thu 17:00–18:00 ET.
- **Holidays:** CME holiday calendar (`exchange_calendars` code `XCME`)

```yaml
market: CME
timezone: America/New_York
weekly_schedule:
  sunday:    [{session: continuous, start: "18:00", end: "00:00+1"}]
  monday:    &cme_weekday
    - {session: continuous, start: "00:00", end: "17:00"}
    - {session: continuous, start: "18:00", end: "00:00+1"}
  tuesday:   *cme_weekday
  wednesday: *cme_weekday
  thursday:  *cme_weekday
  friday:    [{session: continuous, start: "00:00", end: "17:00"}]
  saturday:  []
```

### 4.3 ICE — Brent Crude Oil Futures

- **Timezone:** `America/New_York`
- **Source:** Pyth "Brent Crude (ICE)" section
- **Pattern:** 24/5 with 2-hour daily maintenance window Mon–Thu 18:00–20:00 ET.
- **Holidays:** ICE Futures Europe calendar (`exchange_calendars` code `IFEU`)

```yaml
market: ICE
timezone: America/New_York
weekly_schedule:
  sunday:    [{session: continuous, start: "18:00", end: "00:00+1"}]
  monday:    &ice_weekday
    - {session: continuous, start: "00:00", end: "18:00"}
    - {session: continuous, start: "20:00", end: "00:00+1"}
  tuesday:   *ice_weekday
  wednesday: *ice_weekday
  thursday:  *ice_weekday
  friday:    [{session: continuous, start: "00:00", end: "18:00"}]
  saturday:  []
```

### 4.4 FXCMUKOil — FXCM UKOILSPOT CFD

- **Timezone:** `UTC`
- **Source:** Pyth "UKOILSPOT CFD" entry (FXCM)
- **Pattern:** Mon 01:00–Fri 21:45 UTC. Daily maintenance 22:00–01:00 UTC Mon–Fri (no Sunday session).
- **Holidays:** FXCM broker calendar — hand-curated (approximated as Christmas Day + New Year's Day for 2026)

```yaml
market: FXCMUKOil
timezone: UTC
weekly_schedule:
  sunday:    []
  monday:    &ukoil_weekday [{session: continuous, start: "01:00", end: "22:00"}]
  tuesday:   *ukoil_weekday
  wednesday: *ukoil_weekday
  thursday:  *ukoil_weekday
  friday:    [{session: continuous, start: "01:00", end: "21:45"}]
  saturday:  []
```

### 4.5 FXCMUSOil — FXCM USOILSPOT CFD

- **Timezone:** `UTC`
- **Source:** Pyth "USOILSPOT CFD" entry (FXCM)
- **Pattern:** Sun 23:00 UTC → Fri 21:45 UTC. Daily maintenance 22:00–23:00 UTC Mon–Fri.
- **Holidays:** FXCM broker calendar — hand-curated (same as FXCMUKOil)

```yaml
market: FXCMUSOil
timezone: UTC
weekly_schedule:
  sunday:    [{session: continuous, start: "23:00", end: "00:00+1"}]
  monday:    &usoil_weekday
    - {session: continuous, start: "00:00", end: "22:00"}
    - {session: continuous, start: "23:00", end: "00:00+1"}
  tuesday:   *usoil_weekday
  wednesday: *usoil_weekday
  thursday:  *usoil_weekday
  friday:    [{session: continuous, start: "00:00", end: "21:45"}]
  saturday:  []
```

### 4.6 Rates — Interest Rate Products

- **Timezone:** `America/New_York`
- **Source:** Pyth "Rates" section
- **Pattern:** Mon–Fri 08:00–17:00 ET. Standard weekday business hours — uses `regular` session (not `continuous`).
- **Holidays:** NYSE holiday calendar (`exchange_calendars` code `XNYS`)

```yaml
market: Rates
timezone: America/New_York
weekly_schedule:
  monday:    &rates_weekday [{session: regular, start: "08:00", end: "17:00"}]
  tuesday:   *rates_weekday
  wednesday: *rates_weekday
  thursday:  *rates_weekday
  friday:    *rates_weekday
  saturday:  []
  sunday:    []
```

### 4.7 Metals — Spot Gold/Silver (Au/Ag)

- **Timezone:** `America/New_York`
- **Source:** Pyth "Metals" section
- **Pattern:** Identical to CME: Sun 18:00 ET open, daily 1-hour maintenance Mon–Thu 17:00–18:00 ET, Fri 17:00 ET close.
- **Holidays:** CME holiday calendar (`exchange_calendars` code `XCME`)

```yaml
market: Metals
timezone: America/New_York
weekly_schedule:
  sunday:    [{session: continuous, start: "18:00", end: "00:00+1"}]
  monday:    &metals_weekday
    - {session: continuous, start: "00:00", end: "17:00"}
    - {session: continuous, start: "18:00", end: "00:00+1"}
  tuesday:   *metals_weekday
  wednesday: *metals_weekday
  thursday:  *metals_weekday
  friday:    [{session: continuous, start: "00:00", end: "17:00"}]
  saturday:  []
```

## 5. Holiday Sourcing

| Market | `exchange_calendars` code | Method |
|---|---|---|
| FX | — | Hand-curated: Christmas Day + New Year's Day |
| CME | `XCME` | `exchange_calendars` |
| ICE | `IFEU` | `exchange_calendars` |
| FXCMUKOil | — | Hand-curated: Christmas Day + New Year's Day |
| FXCMUSOil | — | Hand-curated: Christmas Day + New Year's Day |
| Rates | `XNYS` | `exchange_calendars` |
| Metals | `XCME` | `exchange_calendars` |

**`scripts/refresh_holidays.py` additions:**

```python
MARKETS = {
    # existing ...
    "CME":    ("XCME", "cme"),
    "ICE":    ("IFEU", "ice"),
    "Rates":  ("XNYS", "rates"),
    "Metals": ("XCME", "metals"),
}
```

FX and FXCM stay hand-curated: no `exchange_calendars` calendar exists for spot FX or FXCM broker OTC products.

## 6. Files

### Created
- `data/markets/fx.yaml`
- `data/markets/cme.yaml`
- `data/markets/ice.yaml`
- `data/markets/fxcm-ukoil.yaml`
- `data/markets/fxcm-usoil.yaml`
- `data/markets/rates.yaml`
- `data/markets/metals.yaml`
- `data/holidays/fx/2026.yaml`
- `data/holidays/cme/2026.yaml`
- `data/holidays/ice/2026.yaml`
- `data/holidays/fxcm-ukoil/2026.yaml`
- `data/holidays/fxcm-usoil/2026.yaml`
- `data/holidays/rates/2026.yaml`
- `data/holidays/metals/2026.yaml`
- `fx_test.go`
- `cme_test.go`
- `ice_test.go`
- `fxcm_test.go` (covers both FXCMUKOil and FXCMUSOil)
- `rates_test.go`
- `metals_test.go`

### Modified
- `tradinghour.go` — add `SessionContinuous` + 7 `MarketType` constants
- `scripts/refresh_holidays.py` — add 4 new `MARKETS` entries
- `CLAUDE.md` — document new markets in MVP section

### Unchanged
- `loader.go`, `market.go`, `phase.go`, `holiday.go` — no structural changes needed

## 7. Testing Strategy

Each `*_test.go` file contains table-driven tests covering:

1. **Open during active window** — `IsOpen` returns `Open: true, Session: SessionContinuous` (or `SessionRegular` for Rates)
2. **Closed during maintenance gap** — `IsOpen` returns `Open: false, Session: SessionClosed`
3. **Closed on weekend** — Saturday returns closed for all markets
4. **Boundary at open** — exact open time returns open
5. **Boundary at close** — exact close time returns closed (half-open interval semantics)
6. **Timezone correctness** — FX/CME/ICE/Rates/Metals use ET; FXCMUKOil/FXCMUSOil use UTC
7. **Holiday** — a 2026 holiday returns closed

## 8. Out of scope

- Emerging Markets FX (separate Pyth section with per-currency holiday calendars)
- Crypto (24/7, no market hours)
- EU/UK/DE equities (separate from existing equity markets)
- Half-day schedules for any of these markets (none documented on Pyth page)
- FXCM holiday calendar lookup automation (broker proprietary, not in `exchange_calendars`)
