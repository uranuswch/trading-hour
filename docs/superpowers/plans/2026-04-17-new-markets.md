# New Markets Implementation Plan (FX, CME, ICE, FXCM, Rates, Metals)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 7 new `MarketType` constants (FX, CME, ICE, FXCMUKOil, FXCMUSOil, Rates, Metals) to the tradinghour library with correct 24/5 continuous and weekday schedules sourced from the [Pyth Network market hours reference](https://docs.pyth.network/price-feeds/core/market-hours).

**Architecture:** A new `SessionContinuous` constant captures the 24/5 open nature of non-equity markets. Each market gets its own YAML schedule and 2026 holiday file embedded via the existing `go:embed` pipeline. No structural Go changes needed beyond adding constants — the existing loader, market, phase, and holiday packages handle `continuous` as a new string value transparently.

**Tech Stack:** Go 1.22+, `gopkg.in/yaml.v3`, `go:embed`

---

## Codebase Context

Read these before editing anything:

- `tradinghour.go` — public constants (`MarketType`, `Session`) live here
- `loader.go` — `marketDataDir()` maps `MarketType` to `data/holidays/<dir>/`; new markets with simple lowercase names are handled by the existing `default: return strings.ToLower(string(m))` case
- `data/markets/*.yaml` — weekly schedule YAML; `market:` field must equal the `MarketType` string value
- `data/holidays/<dir>/<year>.yaml` — holiday YAML; `market:` field must equal the `MarketType` string value
- Embed glob `data/holidays/*/*.yaml` picks up all new holiday directories automatically
- Existing test helpers `mustNY(t)` (in `isopen_test.go`) and `mustHK(t)` (in `hkex_test.go`) are in-package; reuse `mustNY` for ET markets

## File Structure

**New files (data):**
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
- `data/holidays/fxcmukoil/2026.yaml` ← dir name = `strings.ToLower("FXCMUKOil")`
- `data/holidays/fxcmusoil/2026.yaml` ← dir name = `strings.ToLower("FXCMUSOil")`
- `data/holidays/rates/2026.yaml`
- `data/holidays/metals/2026.yaml`

**New files (tests):**
- `fx_test.go`
- `cme_test.go`
- `ice_test.go`
- `fxcm_test.go`
- `rates_test.go`
- `metals_test.go`

**Modified files:**
- `tradinghour.go` — add `SessionContinuous` + 7 `MarketType` constants
- `scripts/refresh_holidays.py` — add 4 new `MARKETS` entries (CME, ICE, Rates, Metals)
- `CLAUDE.md` — update MVP markets list

---

## Task 1: Add `SessionContinuous` and 7 `MarketType` constants

**Files:**
- Modify: `tradinghour.go`

- [ ] **Step 1: Add `SessionContinuous` to the `Session` const block**

In `tradinghour.go`, the `Session` const block currently ends at `SessionOvernight`. Add one line:

```go
const (
	SessionClosed     Session = "closed"
	SessionPreMarket  Session = "premarket"
	SessionRegular    Session = "regular"
	SessionPostMarket Session = "postmarket"
	SessionOvernight  Session = "overnight"
	SessionContinuous Session = "continuous"
)
```

- [ ] **Step 2: Add 7 new `MarketType` constants**

In `tradinghour.go`, the `MarketType` const block currently ends at `MarketKRX`. Extend it:

```go
const (
	MarketNASDAQ      MarketType = "NASDAQ"
	MarketHKEX        MarketType = "HKEX"
	MarketChinaAShare MarketType = "ChinaAShare"
	MarketTSE         MarketType = "TSE"
	MarketKRX         MarketType = "KRX"
	MarketFX          MarketType = "FX"
	MarketCME         MarketType = "CME"
	MarketICE         MarketType = "ICE"
	MarketFXCMUKOil   MarketType = "FXCMUKOil"
	MarketFXCMUSOil   MarketType = "FXCMUSOil"
	MarketRates       MarketType = "Rates"
	MarketMetals      MarketType = "Metals"
)
```

- [ ] **Step 3: Build and verify no regressions**

```bash
cd /Users/uranuswch/Dev/trading-hour
go build ./...
go test ./...
```

Expected: `ok  github.com/uranuswch/trading-hour` — all 76 existing tests still pass.

- [ ] **Step 4: Commit**

```bash
git add tradinghour.go
git commit -m "feat: add SessionContinuous and 7 new MarketType constants"
```

---

## Task 2: FX — Spot Foreign Exchange

**Source:** Pyth "FX" section — Sun 17:00 ET → Fri 17:00 ET, no maintenance gaps, only Christmas + New Year's closed.

**Files:**
- Create: `fx_test.go`
- Create: `data/markets/fx.yaml`
- Create: `data/holidays/fx/2026.yaml`

- [ ] **Step 1: Write the failing test**

```go
// fx_test.go
package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenFX(t *testing.T) {
	ny := mustNY(t)
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
		wantSess Session
	}{
		// Weekday: fully open 24h
		{"Mon 12:00 open",         time.Date(2026, 3, 2, 12, 0, 0, 0, ny), true,  SessionContinuous},
		{"Mon 00:00 open",         time.Date(2026, 3, 2, 0, 0, 0, 0, ny),  true,  SessionContinuous},
		// Sunday open at 17:00
		{"Sun 16:59 closed",       time.Date(2026, 3, 1, 16, 59, 0, 0, ny), false, SessionClosed},
		{"Sun 17:00 open",         time.Date(2026, 3, 1, 17, 0, 0, 0, ny),  true,  SessionContinuous},
		{"Sun 23:00 open",         time.Date(2026, 3, 1, 23, 0, 0, 0, ny),  true,  SessionContinuous},
		// Friday close at 17:00
		{"Fri 16:59 open",         time.Date(2026, 3, 6, 16, 59, 0, 0, ny), true,  SessionContinuous},
		{"Fri 17:00 closed",       time.Date(2026, 3, 6, 17, 0, 0, 0, ny),  false, SessionClosed},
		// Weekend
		{"Sat 12:00 closed",       time.Date(2026, 3, 7, 12, 0, 0, 0, ny),  false, SessionClosed},
		// Holidays
		{"New Year's Day closed",  time.Date(2026, 1, 1, 12, 0, 0, 0, ny),  false, SessionClosed},
		{"Christmas Day closed",   time.Date(2026, 12, 25, 12, 0, 0, 0, ny), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.local.Unix(), MarketFX)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
cd /Users/uranuswch/Dev/trading-hour
go test -run TestIsOpenFX ./...
```

Expected: FAIL — `tradinghour: unknown market`

- [ ] **Step 3: Create `data/markets/fx.yaml`**

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

- [ ] **Step 4: Create `data/holidays/fx/2026.yaml`**

```yaml
market: FX
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day", type: closed}
  - {date: "2026-12-25", name: "Christmas Day",  type: closed}
```

- [ ] **Step 5: Run to verify it passes**

```bash
go test -run TestIsOpenFX ./...
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add fx_test.go data/markets/fx.yaml data/holidays/fx/2026.yaml
git commit -m "data: add FX (spot forex) schedule, 2026 holidays, and tests"
```

---

## Task 3: CME — WTI Crude Oil Futures

**Source:** Pyth "WTI Crude (CME)" — Sun 18:00 ET → Fri 17:00 ET, daily maintenance Mon–Thu 17:00–18:00 ET.

**Files:**
- Create: `cme_test.go`
- Create: `data/markets/cme.yaml`
- Create: `data/holidays/cme/2026.yaml`

- [ ] **Step 1: Write the failing test**

```go
// cme_test.go
package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenCME(t *testing.T) {
	ny := mustNY(t)
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
		wantSess Session
	}{
		// Normal weekday open
		{"Mon 12:00 open",              time.Date(2026, 3, 2, 12, 0, 0, 0, ny), true,  SessionContinuous},
		// Sunday open at 18:00
		{"Sun 17:59 closed",            time.Date(2026, 3, 1, 17, 59, 0, 0, ny), false, SessionClosed},
		{"Sun 18:00 open",              time.Date(2026, 3, 1, 18, 0, 0, 0, ny),  true,  SessionContinuous},
		// Maintenance window Mon 17:00–18:00
		{"Mon 17:00 closed (maint.)",   time.Date(2026, 3, 2, 17, 0, 0, 0, ny),  false, SessionClosed},
		{"Mon 17:30 closed (maint.)",   time.Date(2026, 3, 2, 17, 30, 0, 0, ny), false, SessionClosed},
		{"Mon 18:00 open (post-maint.)",time.Date(2026, 3, 2, 18, 0, 0, 0, ny),  true,  SessionContinuous},
		// Friday close at 17:00 (no maintenance resumes)
		{"Fri 16:59 open",              time.Date(2026, 3, 6, 16, 59, 0, 0, ny), true,  SessionContinuous},
		{"Fri 17:00 closed",            time.Date(2026, 3, 6, 17, 0, 0, 0, ny),  false, SessionClosed},
		// Weekend
		{"Sat 12:00 closed",            time.Date(2026, 3, 7, 12, 0, 0, 0, ny),  false, SessionClosed},
		// Holiday
		{"Good Friday closed",          time.Date(2026, 4, 3, 12, 0, 0, 0, ny),  false, SessionClosed},
		{"Christmas Day closed",        time.Date(2026, 12, 25, 12, 0, 0, 0, ny), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.local.Unix(), MarketCME)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
go test -run TestIsOpenCME ./...
```

Expected: FAIL — `tradinghour: unknown market`

- [ ] **Step 3: Create `data/markets/cme.yaml`**

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

- [ ] **Step 4: Create `data/holidays/cme/2026.yaml`**

```yaml
market: CME
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",               type: closed}
  - {date: "2026-04-03", name: "Good Friday",                  type: closed}
  - {date: "2026-05-25", name: "Memorial Day",                 type: closed}
  - {date: "2026-07-03", name: "Independence Day (observed)",  type: closed}
  - {date: "2026-09-07", name: "Labor Day",                    type: closed}
  - {date: "2026-11-26", name: "Thanksgiving Day",             type: closed}
  - {date: "2026-12-25", name: "Christmas Day",                type: closed}
```

- [ ] **Step 5: Run to verify it passes**

```bash
go test -run TestIsOpenCME ./...
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add cme_test.go data/markets/cme.yaml data/holidays/cme/2026.yaml
git commit -m "data: add CME (WTI crude) schedule, 2026 holidays, and tests"
```

---

## Task 4: ICE — Brent Crude Oil Futures

**Source:** Pyth "Brent Crude (ICE)" — Sun 18:00 ET → Fri 18:00 ET, daily maintenance Mon–Thu 18:00–20:00 ET. Follows ICE Futures Europe (UK bank holiday calendar).

**Files:**
- Create: `ice_test.go`
- Create: `data/markets/ice.yaml`
- Create: `data/holidays/ice/2026.yaml`

- [ ] **Step 1: Write the failing test**

```go
// ice_test.go
package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenICE(t *testing.T) {
	ny := mustNY(t)
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
		wantSess Session
	}{
		// Normal weekday
		{"Mon 12:00 open",               time.Date(2026, 3, 2, 12, 0, 0, 0, ny), true,  SessionContinuous},
		// Sunday open at 18:00
		{"Sun 17:59 closed",             time.Date(2026, 3, 1, 17, 59, 0, 0, ny), false, SessionClosed},
		{"Sun 18:00 open",               time.Date(2026, 3, 1, 18, 0, 0, 0, ny),  true,  SessionContinuous},
		// Maintenance window Mon 18:00–20:00
		{"Mon 18:00 closed (maint.)",    time.Date(2026, 3, 2, 18, 0, 0, 0, ny),  false, SessionClosed},
		{"Mon 19:59 closed (maint.)",    time.Date(2026, 3, 2, 19, 59, 0, 0, ny), false, SessionClosed},
		{"Mon 20:00 open (post-maint.)", time.Date(2026, 3, 2, 20, 0, 0, 0, ny),  true,  SessionContinuous},
		// Friday close at 18:00
		{"Fri 17:59 open",               time.Date(2026, 3, 6, 17, 59, 0, 0, ny), true,  SessionContinuous},
		{"Fri 18:00 closed",             time.Date(2026, 3, 6, 18, 0, 0, 0, ny),  false, SessionClosed},
		// Weekend
		{"Sat 12:00 closed",             time.Date(2026, 3, 7, 12, 0, 0, 0, ny),  false, SessionClosed},
		// UK bank holidays
		{"Good Friday closed",           time.Date(2026, 4, 3, 12, 0, 0, 0, ny),  false, SessionClosed},
		{"Easter Monday closed",         time.Date(2026, 4, 6, 12, 0, 0, 0, ny),  false, SessionClosed},
		{"Christmas Day closed",         time.Date(2026, 12, 25, 12, 0, 0, 0, ny), false, SessionClosed},
		{"Boxing Day (obs.) closed",     time.Date(2026, 12, 28, 12, 0, 0, 0, ny), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.local.Unix(), MarketICE)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
go test -run TestIsOpenICE ./...
```

Expected: FAIL — `tradinghour: unknown market`

- [ ] **Step 3: Create `data/markets/ice.yaml`**

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

- [ ] **Step 4: Create `data/holidays/ice/2026.yaml`**

ICE Futures Europe observes UK bank holidays.

```yaml
market: ICE
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",          type: closed}
  - {date: "2026-04-03", name: "Good Friday",             type: closed}
  - {date: "2026-04-06", name: "Easter Monday",           type: closed}
  - {date: "2026-05-04", name: "Early May Bank Holiday",  type: closed}
  - {date: "2026-05-25", name: "Spring Bank Holiday",     type: closed}
  - {date: "2026-08-31", name: "Summer Bank Holiday",     type: closed}
  - {date: "2026-12-25", name: "Christmas Day",           type: closed}
  - {date: "2026-12-28", name: "Boxing Day (observed)",   type: closed}
```

- [ ] **Step 5: Run to verify it passes**

```bash
go test -run TestIsOpenICE ./...
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add ice_test.go data/markets/ice.yaml data/holidays/ice/2026.yaml
git commit -m "data: add ICE (Brent crude) schedule, 2026 holidays, and tests"
```

---

## Task 5: FXCM — UK Oil and US Oil CFDs

**Source:** Pyth FXCM entries within Commodities section. Both use `UTC` timezone. FXCM holidays are broker-specific; approximate with Christmas + New Year's.

- **FXCMUKOil (UKOILSPOT):** Mon 01:00 UTC → Fri 21:45 UTC, maintenance 22:00–01:00 UTC daily (Mon–Fri gap). No Sunday session.
- **FXCMUSOil (USOILSPOT):** Sun 23:00 UTC → Fri 21:45 UTC, maintenance 22:00–23:00 UTC daily (Mon–Fri gap).

**Holiday directory names:** `fxcmukoil/` and `fxcmusoil/` (all lowercase, matching `strings.ToLower("FXCMUKOil")` and `strings.ToLower("FXCMUSOil")`)

**Files:**
- Create: `fxcm_test.go`
- Create: `data/markets/fxcm-ukoil.yaml`
- Create: `data/markets/fxcm-usoil.yaml`
- Create: `data/holidays/fxcmukoil/2026.yaml`
- Create: `data/holidays/fxcmusoil/2026.yaml`

- [ ] **Step 1: Write the failing tests**

```go
// fxcm_test.go
package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenFXCMUKOil(t *testing.T) {
	cases := []struct {
		name     string
		utc      time.Time
		wantOpen bool
		wantSess Session
	}{
		// Opens Mon 01:00 UTC
		{"Mon 00:59 closed (before open)", time.Date(2026, 3, 2, 0, 59, 0, 0, time.UTC), false, SessionClosed},
		{"Mon 01:00 open",                 time.Date(2026, 3, 2, 1, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		{"Mon 12:00 open",                 time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC), true,  SessionContinuous},
		// Maintenance 22:00–01:00 UTC
		{"Mon 22:00 closed (maint.)",      time.Date(2026, 3, 2, 22, 0, 0, 0, time.UTC), false, SessionClosed},
		{"Tue 00:59 closed (maint.)",      time.Date(2026, 3, 3, 0, 59, 0, 0, time.UTC), false, SessionClosed},
		{"Tue 01:00 open",                 time.Date(2026, 3, 3, 1, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		// Friday close at 21:45 UTC
		{"Fri 21:44 open",                 time.Date(2026, 3, 6, 21, 44, 0, 0, time.UTC), true,  SessionContinuous},
		{"Fri 21:45 closed",               time.Date(2026, 3, 6, 21, 45, 0, 0, time.UTC), false, SessionClosed},
		// Weekend
		{"Sat 12:00 closed",               time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC), false, SessionClosed},
		{"Sun 12:00 closed",               time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC), false, SessionClosed},
		// Holiday
		{"New Year's Day closed",          time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.utc.Unix(), MarketFXCMUKOil)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}

func TestIsOpenFXCMUSOil(t *testing.T) {
	cases := []struct {
		name     string
		utc      time.Time
		wantOpen bool
		wantSess Session
	}{
		// Opens Sun 23:00 UTC
		{"Sun 22:59 closed",               time.Date(2026, 3, 1, 22, 59, 0, 0, time.UTC), false, SessionClosed},
		{"Sun 23:00 open",                 time.Date(2026, 3, 1, 23, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		{"Mon 12:00 open",                 time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		// Maintenance 22:00–23:00 UTC Mon–Fri
		{"Mon 22:00 closed (maint.)",      time.Date(2026, 3, 2, 22, 0, 0, 0, time.UTC), false, SessionClosed},
		{"Mon 22:59 closed (maint.)",      time.Date(2026, 3, 2, 22, 59, 0, 0, time.UTC), false, SessionClosed},
		{"Mon 23:00 open (post-maint.)",   time.Date(2026, 3, 2, 23, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		// Friday close at 21:45 UTC
		{"Fri 21:44 open",                 time.Date(2026, 3, 6, 21, 44, 0, 0, time.UTC), true,  SessionContinuous},
		{"Fri 21:45 closed",               time.Date(2026, 3, 6, 21, 45, 0, 0, time.UTC), false, SessionClosed},
		// Weekend
		{"Sat 12:00 closed",               time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC), false, SessionClosed},
		// Holiday
		{"New Year's Day closed",          time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.utc.Unix(), MarketFXCMUSOil)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify they fail**

```bash
go test -run 'TestIsOpenFXCM' ./...
```

Expected: FAIL — `tradinghour: unknown market`

- [ ] **Step 3: Create `data/markets/fxcm-ukoil.yaml`**

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

- [ ] **Step 4: Create `data/markets/fxcm-usoil.yaml`**

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

- [ ] **Step 5: Create `data/holidays/fxcmukoil/2026.yaml`**

Note: directory name is `fxcmukoil` (all lowercase), matching `strings.ToLower("FXCMUKOil")`.

```yaml
market: FXCMUKOil
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day", type: closed}
  - {date: "2026-12-25", name: "Christmas Day",  type: closed}
```

- [ ] **Step 6: Create `data/holidays/fxcmusoil/2026.yaml`**

Note: directory name is `fxcmusoil` (all lowercase), matching `strings.ToLower("FXCMUSOil")`.

```yaml
market: FXCMUSOil
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day", type: closed}
  - {date: "2026-12-25", name: "Christmas Day",  type: closed}
```

- [ ] **Step 7: Run to verify both tests pass**

```bash
go test -run 'TestIsOpenFXCM' ./...
```

Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add fxcm_test.go \
  data/markets/fxcm-ukoil.yaml data/markets/fxcm-usoil.yaml \
  data/holidays/fxcmukoil/2026.yaml data/holidays/fxcmusoil/2026.yaml
git commit -m "data: add FXCM UKOil and USOil schedules, 2026 holidays, and tests"
```

---

## Task 6: Rates — Interest Rate Products

**Source:** Pyth "Rates" section — Mon–Fri 08:00–17:00 ET. Standard weekday business hours; uses `regular` session (not `continuous`). Follows NYSE holiday calendar.

**Files:**
- Create: `rates_test.go`
- Create: `data/markets/rates.yaml`
- Create: `data/holidays/rates/2026.yaml`

- [ ] **Step 1: Write the failing test**

```go
// rates_test.go
package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenRates(t *testing.T) {
	ny := mustNY(t)
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
		wantSess Session
	}{
		// Business hours
		{"Mon 07:59 closed",           time.Date(2026, 3, 2, 7, 59, 0, 0, ny),  false, SessionClosed},
		{"Mon 08:00 open",             time.Date(2026, 3, 2, 8, 0, 0, 0, ny),   true,  SessionRegular},
		{"Mon 12:00 open",             time.Date(2026, 3, 2, 12, 0, 0, 0, ny),  true,  SessionRegular},
		{"Mon 16:59 open",             time.Date(2026, 3, 2, 16, 59, 0, 0, ny), true,  SessionRegular},
		{"Mon 17:00 closed",           time.Date(2026, 3, 2, 17, 0, 0, 0, ny),  false, SessionClosed},
		// Weekend
		{"Sat 12:00 closed",           time.Date(2026, 3, 7, 12, 0, 0, 0, ny),  false, SessionClosed},
		{"Sun 12:00 closed",           time.Date(2026, 3, 8, 12, 0, 0, 0, ny),  false, SessionClosed},
		// NYSE holidays
		{"New Year's Day closed",      time.Date(2026, 1, 1, 12, 0, 0, 0, ny),  false, SessionClosed},
		{"MLK Day closed",             time.Date(2026, 1, 19, 12, 0, 0, 0, ny), false, SessionClosed},
		{"Presidents Day closed",      time.Date(2026, 2, 16, 12, 0, 0, 0, ny), false, SessionClosed},
		{"Juneteenth closed",          time.Date(2026, 6, 19, 12, 0, 0, 0, ny), false, SessionClosed},
		{"Thanksgiving closed",        time.Date(2026, 11, 26, 12, 0, 0, 0, ny), false, SessionClosed},
		{"Christmas Day closed",       time.Date(2026, 12, 25, 12, 0, 0, 0, ny), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.local.Unix(), MarketRates)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
go test -run TestIsOpenRates ./...
```

Expected: FAIL — `tradinghour: unknown market`

- [ ] **Step 3: Create `data/markets/rates.yaml`**

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

- [ ] **Step 4: Create `data/holidays/rates/2026.yaml`**

NYSE holidays (same calendar as NASDAQ, minus half-days since Rates has no half-day schedule):

```yaml
market: Rates
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",               type: closed}
  - {date: "2026-01-19", name: "Martin Luther King Jr. Day",   type: closed}
  - {date: "2026-02-16", name: "Presidents' Day",              type: closed}
  - {date: "2026-04-03", name: "Good Friday",                  type: closed}
  - {date: "2026-05-25", name: "Memorial Day",                 type: closed}
  - {date: "2026-06-19", name: "Juneteenth",                   type: closed}
  - {date: "2026-07-03", name: "Independence Day (observed)",  type: closed}
  - {date: "2026-09-07", name: "Labor Day",                    type: closed}
  - {date: "2026-11-26", name: "Thanksgiving Day",             type: closed}
  - {date: "2026-12-25", name: "Christmas Day",                type: closed}
```

- [ ] **Step 5: Run to verify it passes**

```bash
go test -run TestIsOpenRates ./...
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add rates_test.go data/markets/rates.yaml data/holidays/rates/2026.yaml
git commit -m "data: add Rates (interest rates) schedule, 2026 holidays, and tests"
```

---

## Task 7: Metals — Spot Gold/Silver

**Source:** Pyth "Metals" section — Sun 18:00 ET → Fri 17:00 ET, daily maintenance Mon–Thu 17:00–18:00 ET. Identical schedule to CME; follows CME holiday calendar.

**Files:**
- Create: `metals_test.go`
- Create: `data/markets/metals.yaml`
- Create: `data/holidays/metals/2026.yaml`

- [ ] **Step 1: Write the failing test**

```go
// metals_test.go
package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenMetals(t *testing.T) {
	ny := mustNY(t)
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
		wantSess Session
	}{
		// Normal weekday
		{"Mon 12:00 open",               time.Date(2026, 3, 2, 12, 0, 0, 0, ny), true,  SessionContinuous},
		// Sunday open at 18:00
		{"Sun 17:59 closed",             time.Date(2026, 3, 1, 17, 59, 0, 0, ny), false, SessionClosed},
		{"Sun 18:00 open",               time.Date(2026, 3, 1, 18, 0, 0, 0, ny),  true,  SessionContinuous},
		// Maintenance window Mon 17:00–18:00
		{"Mon 17:00 closed (maint.)",    time.Date(2026, 3, 2, 17, 0, 0, 0, ny),  false, SessionClosed},
		{"Mon 17:30 closed (maint.)",    time.Date(2026, 3, 2, 17, 30, 0, 0, ny), false, SessionClosed},
		{"Mon 18:00 open (post-maint.)", time.Date(2026, 3, 2, 18, 0, 0, 0, ny),  true,  SessionContinuous},
		// Friday close at 17:00
		{"Fri 16:59 open",               time.Date(2026, 3, 6, 16, 59, 0, 0, ny), true,  SessionContinuous},
		{"Fri 17:00 closed",             time.Date(2026, 3, 6, 17, 0, 0, 0, ny),  false, SessionClosed},
		// Weekend
		{"Sat 12:00 closed",             time.Date(2026, 3, 7, 12, 0, 0, 0, ny),  false, SessionClosed},
		// CME holidays
		{"Good Friday closed",           time.Date(2026, 4, 3, 12, 0, 0, 0, ny),  false, SessionClosed},
		{"Christmas Day closed",         time.Date(2026, 12, 25, 12, 0, 0, 0, ny), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.local.Unix(), MarketMetals)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
go test -run TestIsOpenMetals ./...
```

Expected: FAIL — `tradinghour: unknown market`

- [ ] **Step 3: Create `data/markets/metals.yaml`**

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

- [ ] **Step 4: Create `data/holidays/metals/2026.yaml`**

```yaml
market: Metals
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",               type: closed}
  - {date: "2026-04-03", name: "Good Friday",                  type: closed}
  - {date: "2026-05-25", name: "Memorial Day",                 type: closed}
  - {date: "2026-07-03", name: "Independence Day (observed)",  type: closed}
  - {date: "2026-09-07", name: "Labor Day",                    type: closed}
  - {date: "2026-11-26", name: "Thanksgiving Day",             type: closed}
  - {date: "2026-12-25", name: "Christmas Day",                type: closed}
```

- [ ] **Step 5: Run to verify it passes**

```bash
go test -run TestIsOpenMetals ./...
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add metals_test.go data/markets/metals.yaml data/holidays/metals/2026.yaml
git commit -m "data: add Metals (spot Au/Ag) schedule, 2026 holidays, and tests"
```

---

## Task 8: Update refresh script and CLAUDE.md

**Files:**
- Modify: `scripts/refresh_holidays.py`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Add 4 new entries to `MARKETS` in `scripts/refresh_holidays.py`**

The current `MARKETS` dict ends at `"KRX": ("XKRX", "krx")`. Extend it:

```python
MARKETS = {
    "NASDAQ":      ("XNAS", "nasdaq"),
    "HKEX":        ("XHKG", "hkex"),
    "ChinaAShare": ("XSHG", "china-ashare"),
    "TSE":         ("XTKS", "tse"),
    "KRX":         ("XKRX", "krx"),
    "CME":         ("XCME", "cme"),
    "ICE":         ("IFEU", "ice"),
    "Rates":       ("XNYS", "rates"),
    "Metals":      ("XCME", "metals"),
}
```

Note: FX and FXCM markets are hand-curated (no `exchange_calendars` calendar). They are intentionally omitted from `MARKETS`.

- [ ] **Step 2: Update `CLAUDE.md` to document the new markets**

Replace the MVP markets section:

```markdown
## MVP markets

**Equity markets:** NASDAQ (including BOAT overnight), HKEX, China A-Share (SSE+SZSE), TSE (Tokyo), KRX (Korea).

**Non-equity markets:** FX (spot forex), CME (WTI crude futures), ICE (Brent crude futures), FXCMUKOil (UKOILSPOT CFD), FXCMUSOil (USOILSPOT CFD), Rates (interest rate products), Metals (spot Au/Ag).

Non-equity markets use `SessionContinuous` for open windows; maintenance gaps appear as closed periods. FX and FXCM holiday calendars are hand-curated (no `exchange_calendars` support).
```

- [ ] **Step 3: Run full test suite**

```bash
go test -race ./...
```

Expected: all tests pass (should now be ~96+ tests).

- [ ] **Step 4: Commit**

```bash
git add scripts/refresh_holidays.py CLAUDE.md
git commit -m "chore: update refresh script and CLAUDE.md for new non-equity markets"
```

---

## Final Verification

- [ ] `go test -race ./...` — all pass
- [ ] `go vet ./...` — clean
- [ ] `go build ./...` — clean
- [ ] `find data -type f | sort` — 14 market YAMLs + 14 holiday YAMLs (7 new each, plus 2 `.gitkeep` files)

---

## Spec Coverage

| Spec section | Covered by tasks |
|---|---|
| New session type `SessionContinuous` | Task 1 |
| 7 new `MarketType` constants | Task 1 |
| FX schedule + holidays | Task 2 |
| CME schedule + holidays | Task 3 |
| ICE schedule + holidays | Task 4 |
| FXCMUKOil + FXCMUSOil schedules + holidays | Task 5 |
| Rates schedule + holidays | Task 6 |
| Metals schedule + holidays | Task 7 |
| `scripts/refresh_holidays.py` update | Task 8 |
| `CLAUDE.md` update | Task 8 |

## Out of scope

- Emerging Markets FX (separate per-currency holiday calendars)
- Crypto markets (24/7, no market hours concept)
- EU/UK/DE equities
- Half-day schedules (none documented on Pyth page for these markets)
- FXCM holiday calendar automation (broker proprietary)
