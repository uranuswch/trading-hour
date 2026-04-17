# Rates and Metals Markets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the new markets feature by implementing Rates (interest rate products) and Metals (spot gold/silver) markets with full data, holidays, and test coverage.

**Architecture:** Two new markets following the established pattern from FX, CME, ICE, and FXCM implementations. Rates uses discrete `SessionRegular` hours; Metals uses `SessionContinuous` with daily maintenance gaps identical to CME. Both use `America/New_York` timezone and source holidays from `exchange_calendars`.

**Tech Stack:** Go 1.23+, YAML data files, table-driven Go tests, Python holidays script

---

## File Structure

**Created:**
- `data/markets/rates.yaml` - Rates weekly schedule (Mon-Fri 08:00-17:00 ET)
- `data/markets/metals.yaml` - Metals weekly schedule (24/5 with 1-hour maintenance)
- `data/holidays/rates/2026.yaml` - Rates 2026 holidays (NYSE calendar)
- `data/holidays/metals/2026.yaml` - Metals 2026 holidays (CME calendar)
- `rates_test.go` - Table-driven tests for Rates market
- `metals_test.go` - Table-driven tests for Metals market

**Modified:**
- `scripts/refresh_holidays.py` - Add CME, ICE, Rates, Metals to MARKETS dict
- `CLAUDE.md` - Document Rates and Metals in MVP markets section

---

## Task 1: Add Missing Markets to Holiday Refresh Script

**Files:**
- Modify: `scripts/refresh_holidays.py:19-25`

- [ ] **Step 1: Update MARKETS dictionary in refresh_holidays.py**

Replace the existing MARKETS dict to include CME, ICE, Rates, and Metals:

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

- [ ] **Step 2: Verify the script runs without errors**

Run: `python3 scripts/refresh_holidays.py --help`
Expected: Script shows help message without syntax errors

- [ ] **Step 3: Commit**

```bash
git add scripts/refresh_holidays.py
git commit -m "feat: add CME, ICE, Rates, Metals to holiday refresh script"
```

---

## Task 2: Generate Rates 2026 Holidays

**Files:**
- Create: `data/holidays/rates/2026.yaml`

- [ ] **Step 1: Run holiday generation script for Rates**

Run: `python3 scripts/refresh_holidays.py Rates 2026`
Expected: Script generates `data/holidays/rates/2026.yaml` with NYSE holidays

- [ ] **Step 2: Verify the generated file exists and has content**

Run: `head -20 data/holidays/rates/2026.yaml`
Expected: YAML file with `year: 2026` and holiday entries

- [ ] **Step 3: Manually verify key holidays are present**

Run: `grep -E "(New Year's Day|Thanksgiving|Christmas) data/holidays/rates/2026.yaml`
Expected: At least these major NYSE holidays present for 2026

- [ ] **Step 4: Commit**

```bash
git add data/holidays/rates/2026.yaml
git commit -m "data: add Rates (NYSE) 2026 holidays"
```

---

## Task 3: Generate Metals 2026 Holidays

**Files:**
- Create: `data/holidays/metals/2026.yaml`

- [ ] **Step 1: Run holiday generation script for Metals**

Run: `python3 scripts/refresh_holidays.py Metals 2026`
Expected: Script generates `data/holidays/metals/2026.yaml` with CME holidays

- [ ] **Step 2: Verify the generated file exists and has content**

Run: `head -20 data/holidays/metals/2026.yaml`
Expected: YAML file with `year: 2026` and holiday entries

- [ ] **Step 3: Manually verify key holidays are present**

Run: `grep -E "(New Year's Day|Independence Day|Christmas) data/holidays/metals/2026.yaml`
Expected: At least these major CME holidays present for 2026

- [ ] **Step 4: Commit**

```bash
git add data/holidays/metals/2026.yaml
git commit -m "data: add Metals (CME) 2026 holidays"
```

---

## Task 4: Create Rates Market Schedule

**Files:**
- Create: `data/markets/rates.yaml`

- [ ] **Step 1: Create rates.yaml with weekly schedule**

Create file with this content:

```yaml
market: Rates
timezone: America/New_York
weekly_schedule:
  monday:    &rates_day [{session: regular, start: "08:00", end: "17:00"}]
  tuesday:   *rates_day
  wednesday: *rates_day
  thursday:  *rates_day
  friday:    *rates_day
  saturday:  []
  sunday:    []
```

- [ ] **Step 2: Verify YAML syntax is valid**

Run: `python3 -c "import yaml; yaml.safe_load(open('data/markets/rates.yaml'))"`
Expected: No syntax errors

- [ ] **Step 3: Commit**

```bash
git add data/markets/rates.yaml
git commit -m "data: add Rates (interest rate products) schedule"
```

---

## Task 5: Create Metals Market Schedule

**Files:**
- Create: `data/markets/metals.yaml`

- [ ] **Step 1: Create metals.yaml with weekly schedule**

Create file with this content:

```yaml
market: Metals
timezone: America/New_York
weekly_schedule:
  sunday:    [{session: continuous, start: "18:00", end: "00:00+1"}]
  monday:    &metals_day
    - {session: continuous, start: "00:00", end: "17:00"}
    - {session: continuous, start: "18:00", end: "00:00+1"}
  tuesday:   *metals_day
  wednesday: *metals_day
  thursday:  *metals_day
  friday:    [{session: continuous, start: "00:00", end: "17:00"}]
  saturday:  []
```

- [ ] **Step 2: Verify YAML syntax is valid**

Run: `python3 -c "import yaml; yaml.safe_load(open('data/markets/metals.yaml'))"`
Expected: No syntax errors

- [ ] **Step 3: Commit**

```bash
git add data/markets/metals.yaml
git commit -m "data: add Metals (spot gold/silver) schedule"
```

---

## Task 6: Write Rates Tests

**Files:**
- Create: `rates_test.go`

- [ ] **Step 1: Create rates_test.go with table-driven tests**

Create file with this content:

```go
package tradinghour

import (
	"testing"
	"time"
)

func TestRates(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name        string
		timestamp   int64
		wantOpen    bool
		wantSession Session
	}{
		{
			name:        "open during regular hours",
			timestamp:   time.Date(2026, 1, 5, 10, 0, 0, 0, ny).Unix(), // Monday 10am ET
			wantOpen:    true,
			wantSession: SessionRegular,
		},
		{
			name:        "closed before open",
			timestamp:   time.Date(2026, 1, 5, 7, 0, 0, 0, ny).Unix(), // Monday 7am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "closed after close",
			timestamp:   time.Date(2026, 1, 5, 18, 0, 0, 0, ny).Unix(), // Monday 6pm ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "closed on Saturday",
			timestamp:   time.Date(2026, 1, 3, 10, 0, 0, 0, ny).Unix(), // Saturday 10am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "closed on Sunday",
			timestamp:   time.Date(2026, 1, 4, 10, 0, 0, 0, ny).Unix(), // Sunday 10am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "boundary at open",
			timestamp:   time.Date(2026, 1, 5, 8, 0, 0, 0, ny).Unix(), // Monday 8:00:00 ET
			wantOpen:    true,
			wantSession: SessionRegular,
		},
		{
			name:        "boundary at close",
			timestamp:   time.Date(2026, 1, 5, 17, 0, 0, 0, ny).Unix(), // Monday 17:00:00 ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "closed on New Year's Day 2026",
			timestamp:   time.Date(2026, 1, 1, 10, 0, 0, 0, ny).Unix(), // Jan 1 10am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOpen(tt.timestamp, MarketRates)
			if got.Open != tt.wantOpen {
				t.Errorf("IsOpen() Open = %v, want %v", got.Open, tt.wantOpen)
			}
			if got.Session != tt.wantSession {
				t.Errorf("IsOpen() Session = %v, want %v", got.Session, tt.wantSession)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail (no implementation yet)**

Run: `go test -v -run TestRates`
Expected: Tests pass (loader should auto-load the new market)

- [ ] **Step 3: Commit**

```bash
git add rates_test.go
git commit -m "test: add Rates market table-driven tests"
```

---

## Task 7: Write Metals Tests

**Files:**
- Create: `metals_test.go`

- [ ] **Step 1: Create metals_test.go with table-driven tests**

Create file with this content:

```go
package tradinghour

import (
	"testing"
	"time"
)

func TestMetals(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name        string
		timestamp   int64
		wantOpen    bool
		wantSession Session
	}{
		{
			name:        "open during continuous session",
			timestamp:   time.Date(2026, 1, 5, 10, 0, 0, 0, ny).Unix(), // Monday 10am ET
			wantOpen:    true,
			wantSession: SessionContinuous,
		},
		{
			name:        "closed during maintenance gap",
			timestamp:   time.Date(2026, 1, 5, 17, 30, 0, 0, ny).Unix(), // Monday 5:30pm ET (in 17:00-18:00 gap)
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "closed on Saturday",
			timestamp:   time.Date(2026, 1, 3, 10, 0, 0, 0, ny).Unix(), // Saturday 10am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "open Sunday evening",
			timestamp:   time.Date(2026, 1, 4, 19, 0, 0, 0, ny).Unix(), // Sunday 7pm ET
			wantOpen:    true,
			wantSession: SessionContinuous,
		},
		{
			name:        "boundary at Sunday open",
			timestamp:   time.Date(2026, 1, 4, 18, 0, 0, 0, ny).Unix(), // Sunday 18:00:00 ET
			wantOpen:    true,
			wantSession: SessionContinuous,
		},
		{
			name:        "boundary at Friday close",
			timestamp:   time.Date(2026, 1, 9, 17, 0, 0, 0, ny).Unix(), // Friday 17:00:00 ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "open after maintenance ends",
			timestamp:   time.Date(2026, 1, 5, 18, 0, 0, 0, ny).Unix(), // Monday 18:00:00 ET
			wantOpen:    true,
			wantSession: SessionContinuous,
		},
		{
			name:        "closed on Christmas 2026",
			timestamp:   time.Date(2026, 12, 25, 10, 0, 0, 0, ny).Unix(), // Dec 25 10am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOpen(tt.timestamp, MarketMetals)
			if got.Open != tt.wantOpen {
				t.Errorf("IsOpen() Open = %v, want %v", got.Open, tt.wantOpen)
			}
			if got.Session != tt.wantSession {
				t.Errorf("IsOpen() Session = %v, want %v", got.Session, tt.wantSession)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `go test -v -run TestMetals`
Expected: Tests pass (loader should auto-load the new market)

- [ ] **Step 3: Commit**

```bash
git add metals_test.go
git commit -m "test: add Metals market table-driven tests"
```

---

## Task 8: Update CLAUDE.md Documentation

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update MVP markets section in CLAUDE.md**

Read the current MVP section and add Rates and Metals to the list. The section currently lists NASDAQ, HKEX, China A-Share, TSE, KRX. Add: "Rates (interest rate products), Metals (spot gold/silver)."

After the existing MVP markets list, add:

```
- Rates (interest rate products)
- Metals (spot gold/silver)
```

- [ ] **Step 2: Verify documentation is clear and accurate**

Read: `head -30 CLAUDE.md`
Expected: MVP section lists all 7 new markets

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add Rates and Metals to MVP markets list"
```

---

## Task 9: Final Verification

**Files:**
- Test all changes

- [ ] **Step 1: Run all tests to ensure nothing is broken**

Run: `go test -v ./...`
Expected: All tests pass, including existing FX, CME, ICE, FXCM tests and new Rates, Metals tests

- [ ] **Step 2: Verify data files load correctly**

Run: `go run -c 'package main; import "fmt"; import "time"; import "tradinghour"; func main() { ny, _ := time.LoadLocation("America/New_York"); fmt.Println(tradinghour.IsOpen(time.Date(2026, 1, 5, 10, 0, 0, 0, ny).Unix(), tradinghour.MarketRates)) }'`

Or simpler: Create a tiny test program and run it

Expected: No errors, markets load successfully

- [ ] **Step 3: Verify all 7 new markets have data**

Run: `ls -1 data/markets/*.yaml | wc -l`
Expected: At least 12 (5 original + 7 new)

Run: `ls -1 data/holidays/ | grep -E "(fx|cme|ice|fxcm|rates|metals)"`
Expected: 7 directories (fx, cme, ice, fxcmukoil, fxcmusoil, rates, metals)

- [ ] **Step 4: Check git status**

Run: `git status`
Expected: Clean working directory (all changes committed)

---

## Self-Review Checklist

- [ ] **Spec coverage:** All 7 markets from design spec are implemented (FX, CME, ICE, FXCMUKOil, FXCMUSOil, Rates, Metals)
- [ ] **Data files exist:** All 7 market YAML files created
- [ ] **Holiday data exists:** All 7 markets have 2026 holiday files
- [ ] **Tests exist:** All 7 markets have test files (fxcm_test.go covers both FXCM markets)
- [ ] **Script updated:** refresh_holidays.py includes CME, ICE, Rates, Metals
- [ ] **Documentation updated:** CLAUDE.md lists all new markets
- [ ] **No placeholders:** Every step has complete code/commands
- [ ] **Type consistency:** SessionContinuous used for Metals, SessionRegular used for Rates
- [ ] **Timezone correctness:** Rates and Metals both use America/New_York
- [ ] **All tests pass:** No test failures
