# `tradinghour` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go library that, given `(unix_ts, market)` returns open-state + session, and given `(date, market)` returns a full day timeline — with embedded YAML schedule/holiday data for NASDAQ (incl. BOAT overnight), HKEX, China A-Share, TSE, and KRX, plus a yearly GitHub Action to refresh holidays.

**Architecture:** Flat Go package (`tradinghour`). YAML data files embedded via `go:embed`, parsed once at package init into compiled in-memory structs. All computations anchored in the market's local timezone via `time.LoadLocation`. Public API: `IsOpen`, `Timeline`, `NextOpen`, `NextClose`. Yearly Python script generates next-year holiday YAML from `exchange_calendars` and opens a PR.

**Tech Stack:** Go 1.22+, `gopkg.in/yaml.v3`, Python 3.11 + `exchange_calendars` (GH Action only), `peter-evans/create-pull-request@v6`.

**Spec:** [docs/superpowers/specs/2026-04-17-tradinghour-design.md](../specs/2026-04-17-tradinghour-design.md)

---

## File map

Create:
- `go.mod`, `go.sum`
- `tradinghour.go` — public types (`MarketType`, `Session`, `Status`, `Phase`, `DaySchedule`, `ErrUnknownMarket`) + top-level functions (`IsOpen`, `Timeline`, `NextOpen`, `NextClose`).
- `phase.go` — `timeOfDay` parsing (`"HH:MM"` / `"HH:MM+1"`), `compiledPhase`, instantiation to `time.Time`.
- `market.go` — `Market` struct, `materialize(date)` returning `(phases, isHoliday, isHalfDay, name)`.
- `holiday.go` — `HolidayType`, `holidayEntry`, `civilDate` key.
- `loader.go` — `go:embed` filesystem, YAML parsing, `init()` wiring all markets into a registry.
- `data/markets/{nasdaq,hkex,china-ashare,tse,krx}.yaml`
- `data/holidays/{nasdaq,hkex,china-ashare,tse,krx}/2026.yaml`
- `tradinghour_test.go`, `phase_test.go`, `market_test.go`, per-market test files.
- `scripts/refresh_holidays.py`
- `.github/workflows/refresh-holidays.yml`
- `README.md`
- `.gitignore`

---

## Task 1: Bootstrap Go module and directory structure

**Files:**
- Create: `go.mod`, `.gitignore`, `tradinghour.go`, `data/markets/.gitkeep`, `data/holidays/.gitkeep`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
cd /Users/uranuswch/Dev/trading-hour
go mod init github.com/uranuswch/trading-hour
go get gopkg.in/yaml.v3
```

Expected: `go.mod` created with `module github.com/uranuswch/trading-hour` and yaml.v3 dependency.

- [ ] **Step 2: Create `.gitignore`**

Create `.gitignore`:
```
# Binaries
*.exe
*.test
*.out

# Editor
.idea/
.vscode/
*.swp

# Local env
.env
```

- [ ] **Step 3: Create empty package file so `go build` succeeds**

Create `tradinghour.go`:
```go
// Package tradinghour provides trading-hours queries for global equity markets.
package tradinghour
```

- [ ] **Step 4: Scaffold data directories**

Run:
```bash
mkdir -p data/markets data/holidays
touch data/markets/.gitkeep data/holidays/.gitkeep
```

- [ ] **Step 5: Verify build**

Run: `go build ./...`
Expected: no output, exit 0.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum .gitignore tradinghour.go data/
git commit -m "chore: bootstrap Go module and package skeleton"
```

---

## Task 2: Public types and error sentinel

**Files:**
- Modify: `tradinghour.go`
- Create: `tradinghour_test.go`

- [ ] **Step 1: Write failing test for type existence**

Create `tradinghour_test.go`:
```go
package tradinghour

import (
	"errors"
	"testing"
)

func TestMarketTypeConstants(t *testing.T) {
	want := map[MarketType]string{
		MarketNASDAQ:      "NASDAQ",
		MarketHKEX:        "HKEX",
		MarketChinaAShare: "ChinaAShare",
		MarketTSE:         "TSE",
		MarketKRX:         "KRX",
	}
	for got, s := range want {
		if string(got) != s {
			t.Errorf("MarketType %q != %q", got, s)
		}
	}
}

func TestSessionConstants(t *testing.T) {
	for _, s := range []Session{
		SessionClosed, SessionPreMarket, SessionRegular,
		SessionPostMarket, SessionOvernight,
	} {
		if s == "" {
			t.Errorf("session constant is empty")
		}
	}
}

func TestErrUnknownMarket(t *testing.T) {
	if !errors.Is(ErrUnknownMarket, ErrUnknownMarket) {
		t.Fatal("ErrUnknownMarket sentinel must be comparable with errors.Is")
	}
}
```

- [ ] **Step 2: Run test — expect build failure**

Run: `go test ./...`
Expected: compile errors (undefined identifiers).

- [ ] **Step 3: Implement types in `tradinghour.go`**

Replace `tradinghour.go` contents with:
```go
// Package tradinghour provides trading-hours queries for global equity markets.
package tradinghour

import (
	"errors"
	"time"
)

// MarketType identifies a supported market.
type MarketType string

const (
	MarketNASDAQ      MarketType = "NASDAQ"
	MarketHKEX        MarketType = "HKEX"
	MarketChinaAShare MarketType = "ChinaAShare"
	MarketTSE         MarketType = "TSE"
	MarketKRX         MarketType = "KRX"
)

// Session identifies a market phase at a given instant.
type Session string

const (
	SessionClosed     Session = "closed"
	SessionPreMarket  Session = "premarket"
	SessionRegular    Session = "regular"
	SessionPostMarket Session = "postmarket"
	SessionOvernight  Session = "overnight"
)

// Status is the result of IsOpen.
type Status struct {
	Open    bool
	Session Session
	Market  MarketType
}

// Phase is one open interval on a given day. End may fall on a later calendar day (overnight).
type Phase struct {
	Session Session
	Start   time.Time
	End     time.Time
}

// DaySchedule is the full timeline for a single market-local date.
type DaySchedule struct {
	Date        time.Time
	Market      MarketType
	Phases      []Phase
	IsHoliday   bool
	IsHalfDay   bool
	HolidayName string
}

// ErrUnknownMarket is returned for any API call with an unknown MarketType.
var ErrUnknownMarket = errors.New("tradinghour: unknown market")

// Placeholder stubs so the API surface exists; real implementations come in later tasks.
func IsOpen(unixSec int64, m MarketType) (Status, error)          { panic("not implemented") }
func Timeline(date time.Time, m MarketType) (DaySchedule, error)  { panic("not implemented") }
func NextOpen(unixSec int64, m MarketType) (time.Time, error)     { panic("not implemented") }
func NextClose(unixSec int64, m MarketType) (time.Time, error)    { panic("not implemented") }
```

- [ ] **Step 4: Run test — expect pass**

Run: `go test -run 'TestMarketTypeConstants|TestSessionConstants|TestErrUnknownMarket' ./...`
Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add tradinghour.go tradinghour_test.go
git commit -m "feat: add public types and ErrUnknownMarket sentinel"
```

---

## Task 3: `timeOfDay` parsing and phase instantiation

**Files:**
- Create: `phase.go`, `phase_test.go`

- [ ] **Step 1: Write failing tests**

Create `phase_test.go`:
```go
package tradinghour

import (
	"testing"
	"time"
)

func TestParseTimeOfDay(t *testing.T) {
	cases := []struct {
		in        string
		h, m, off int
		wantErr   bool
	}{
		{"04:00", 4, 0, 0, false},
		{"09:30", 9, 30, 0, false},
		{"16:00", 16, 0, 0, false},
		{"23:59", 23, 59, 0, false},
		{"04:00+1", 4, 0, 1, false},
		{"00:00+1", 0, 0, 1, false},
		{"", 0, 0, 0, true},
		{"9:30", 0, 0, 0, true},     // need zero-padded HH
		{"25:00", 0, 0, 0, true},
		{"10:60", 0, 0, 0, true},
		{"10:30+2", 0, 0, 0, true},  // only +1 allowed in MVP
	}
	for _, c := range cases {
		got, err := parseTimeOfDay(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseTimeOfDay(%q) expected error, got %+v", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseTimeOfDay(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got.hour != c.h || got.minute != c.m || got.dayOffset != c.off {
			t.Errorf("parseTimeOfDay(%q) = %+v, want {%d %d %d}", c.in, got, c.h, c.m, c.off)
		}
	}
}

func TestInstantiatePhase(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(2026, 3, 2, 0, 0, 0, 0, loc) // Mon
	cp := compiledPhase{
		Session: SessionOvernight,
		Start:   timeOfDay{20, 0, 0},
		End:     timeOfDay{4, 0, 1},
	}
	p := cp.instantiate(date, loc)
	if p.Start != time.Date(2026, 3, 2, 20, 0, 0, 0, loc) {
		t.Errorf("Start = %v", p.Start)
	}
	if p.End != time.Date(2026, 3, 3, 4, 0, 0, 0, loc) {
		t.Errorf("End = %v", p.End)
	}
	if p.Session != SessionOvernight {
		t.Errorf("Session = %v", p.Session)
	}
}

func TestInstantiatePhaseDSTSpringForward(t *testing.T) {
	// 2026-03-08 is US spring-forward. An overnight phase starting Sun 20:00
	// should end at Mon 04:00 local, crossing the DST boundary correctly.
	loc, _ := time.LoadLocation("America/New_York")
	date := time.Date(2026, 3, 8, 0, 0, 0, 0, loc)
	cp := compiledPhase{
		Session: SessionOvernight,
		Start:   timeOfDay{20, 0, 0},
		End:     timeOfDay{4, 0, 1},
	}
	p := cp.instantiate(date, loc)
	want := time.Date(2026, 3, 9, 4, 0, 0, 0, loc)
	if !p.End.Equal(want) {
		t.Errorf("DST End = %v, want %v", p.End, want)
	}
}
```

- [ ] **Step 2: Run tests — expect build failure**

Run: `go test -run 'TestParseTimeOfDay|TestInstantiatePhase' ./...`
Expected: compile errors.

- [ ] **Step 3: Implement `phase.go`**

Create `phase.go`:
```go
package tradinghour

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// timeOfDay is an hour/minute plus an optional +1-day offset, used for phases
// that end on the next calendar day (e.g. NASDAQ BOAT 20:00 -> 04:00+1).
type timeOfDay struct {
	hour, minute, dayOffset int
}

// parseTimeOfDay parses "HH:MM" or "HH:MM+1".
func parseTimeOfDay(s string) (timeOfDay, error) {
	var tod timeOfDay
	if s == "" {
		return tod, fmt.Errorf("tradinghour: empty time string")
	}
	base := s
	if strings.HasSuffix(s, "+1") {
		tod.dayOffset = 1
		base = strings.TrimSuffix(s, "+1")
	}
	if len(base) != 5 || base[2] != ':' {
		return tod, fmt.Errorf("tradinghour: invalid time %q (need HH:MM)", s)
	}
	h, err := strconv.Atoi(base[0:2])
	if err != nil || h < 0 || h > 23 {
		return tod, fmt.Errorf("tradinghour: invalid hour in %q", s)
	}
	m, err := strconv.Atoi(base[3:5])
	if err != nil || m < 0 || m > 59 {
		return tod, fmt.Errorf("tradinghour: invalid minute in %q", s)
	}
	tod.hour = h
	tod.minute = m
	return tod, nil
}

// compiledPhase is a phase with parsed times, ready to be instantiated onto a specific date.
type compiledPhase struct {
	Session Session
	Start   timeOfDay
	End     timeOfDay
}

// instantiate materializes this phase onto the given date in the given location.
// The date argument must already be market-local-midnight.
func (c compiledPhase) instantiate(date time.Time, loc *time.Location) Phase {
	y, m, d := date.Date()
	return Phase{
		Session: c.Session,
		Start:   time.Date(y, m, d+c.Start.dayOffset, c.Start.hour, c.Start.minute, 0, 0, loc),
		End:     time.Date(y, m, d+c.End.dayOffset, c.End.hour, c.End.minute, 0, 0, loc),
	}
}
```

- [ ] **Step 4: Run tests — expect pass**

Run: `go test -run 'TestParseTimeOfDay|TestInstantiatePhase' ./...`
Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add phase.go phase_test.go
git commit -m "feat: add timeOfDay parsing and phase instantiation"
```

---

## Task 4: Holiday types and lookup

**Files:**
- Create: `holiday.go`

- [ ] **Step 1: Create `holiday.go`**

No tests yet — these are pure data types used by the loader in Task 5.

Create `holiday.go`:
```go
package tradinghour

// HolidayType distinguishes full-closure days from early-close days.
type HolidayType string

const (
	HolidayClosed  HolidayType = "closed"
	HolidayHalfDay HolidayType = "half_day"
)

// civilDate is a timezone-agnostic Y/M/D used as a map key for holiday lookup.
type civilDate struct {
	year  int
	month int // 1..12
	day   int // 1..31
}

// holidayEntry describes a single holiday date.
type holidayEntry struct {
	Name string
	Type HolidayType
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add holiday.go
git commit -m "feat: add holiday data types"
```

---

## Task 5: NASDAQ schedule YAML file

**Files:**
- Create: `data/markets/nasdaq.yaml`

- [ ] **Step 1: Create `data/markets/nasdaq.yaml`**

Based on spec section 4.1:
```yaml
market: NASDAQ
timezone: America/New_York
weekly_schedule:
  monday: &weekday_full
    - {session: premarket,  start: "04:00", end: "09:30"}
    - {session: regular,    start: "09:30", end: "16:00"}
    - {session: postmarket, start: "16:00", end: "20:00"}
    - {session: overnight,  start: "20:00", end: "04:00+1"}
  tuesday: *weekday_full
  wednesday: *weekday_full
  thursday: *weekday_full
  friday:
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

- [ ] **Step 2: Commit**

```bash
git add data/markets/nasdaq.yaml
git commit -m "data: add NASDAQ weekly schedule with BOAT overnight"
```

---

## Task 6: NASDAQ 2026 holidays YAML

**Files:**
- Create: `data/holidays/nasdaq/2026.yaml`

- [ ] **Step 1: Create `data/holidays/nasdaq/2026.yaml`**

Source: [NYSE calendar](https://www.nyse.com/markets/hours-calendars) for 2026.

```yaml
market: NASDAQ
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",                 type: closed}
  - {date: "2026-01-19", name: "Martin Luther King Jr. Day",     type: closed}
  - {date: "2026-02-16", name: "Presidents' Day",                type: closed}
  - {date: "2026-04-03", name: "Good Friday",                    type: closed}
  - {date: "2026-05-25", name: "Memorial Day",                   type: closed}
  - {date: "2026-06-19", name: "Juneteenth",                     type: closed}
  - {date: "2026-07-03", name: "Independence Day (observed)",    type: closed}
  - {date: "2026-09-07", name: "Labor Day",                      type: closed}
  - {date: "2026-11-26", name: "Thanksgiving Day",               type: closed}
  - {date: "2026-11-27", name: "Day after Thanksgiving",         type: half_day}
  - {date: "2026-12-24", name: "Christmas Eve",                  type: half_day}
  - {date: "2026-12-25", name: "Christmas Day",                  type: closed}
```

- [ ] **Step 2: Commit**

```bash
git add data/holidays/nasdaq/2026.yaml
git commit -m "data: add NASDAQ 2026 holiday calendar"
```

---

## Task 7: Loader — embed + parse market + holidays into a registry

**Files:**
- Create: `loader.go`, `market.go`

- [ ] **Step 1: Create `market.go` with the `Market` struct**

```go
package tradinghour

import "time"

// Market holds all compiled schedule and holiday data for a single market.
type Market struct {
	Type          MarketType
	Location      *time.Location
	WeeklyPhases  [7][]compiledPhase // index = int(time.Weekday); Sunday = 0
	HalfDayPhases []compiledPhase    // nil if market has no half-day support
	Holidays      map[civilDate]holidayEntry
}

// registry is populated by loader.go's init().
var registry = map[MarketType]*Market{}

// lookup returns the registered market or an error.
func lookup(m MarketType) (*Market, error) {
	mkt, ok := registry[m]
	if !ok {
		return nil, ErrUnknownMarket
	}
	return mkt, nil
}
```

- [ ] **Step 2: Create `loader.go` with embed + parsing**

```go
package tradinghour

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed data/markets/*.yaml data/holidays/*/*.yaml
var dataFS embed.FS

type scheduleFile struct {
	Market          string                 `yaml:"market"`
	Timezone        string                 `yaml:"timezone"`
	WeeklySchedule  map[string][]phaseYAML `yaml:"weekly_schedule"`
	HalfDaySchedule []phaseYAML            `yaml:"half_day_schedule"`
}

type phaseYAML struct {
	Session string `yaml:"session"`
	Start   string `yaml:"start"`
	End     string `yaml:"end"`
}

type holidayFile struct {
	Market   string         `yaml:"market"`
	Year     int            `yaml:"year"`
	Holidays []holidayYAML  `yaml:"holidays"`
}

type holidayYAML struct {
	Date string `yaml:"date"`
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

var weekdayByName = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
	"wednesday": time.Wednesday, "thursday": time.Thursday,
	"friday": time.Friday, "saturday": time.Saturday,
}

func init() {
	entries, err := fs.ReadDir(dataFS, "data/markets")
	if err != nil {
		panic(fmt.Errorf("tradinghour: read markets dir: %w", err))
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		if err := loadMarket("data/markets/" + e.Name()); err != nil {
			panic(err)
		}
	}
	if err := loadHolidays(); err != nil {
		panic(err)
	}
}

func loadMarket(p string) error {
	raw, err := dataFS.ReadFile(p)
	if err != nil {
		return fmt.Errorf("tradinghour: read %s: %w", p, err)
	}
	var sf scheduleFile
	if err := yaml.Unmarshal(raw, &sf); err != nil {
		return fmt.Errorf("tradinghour: parse %s: %w", p, err)
	}
	loc, err := time.LoadLocation(sf.Timezone)
	if err != nil {
		return fmt.Errorf("tradinghour: load tz %q: %w", sf.Timezone, err)
	}
	m := &Market{
		Type:     MarketType(sf.Market),
		Location: loc,
		Holidays: map[civilDate]holidayEntry{},
	}
	for dayName, phases := range sf.WeeklySchedule {
		wd, ok := weekdayByName[strings.ToLower(dayName)]
		if !ok {
			return fmt.Errorf("tradinghour: unknown weekday %q in %s", dayName, p)
		}
		cps, err := compilePhases(phases, p)
		if err != nil {
			return err
		}
		m.WeeklyPhases[int(wd)] = cps
	}
	if len(sf.HalfDaySchedule) > 0 {
		cps, err := compilePhases(sf.HalfDaySchedule, p)
		if err != nil {
			return err
		}
		m.HalfDayPhases = cps
	}
	registry[m.Type] = m
	return nil
}

func compilePhases(ps []phaseYAML, srcPath string) ([]compiledPhase, error) {
	out := make([]compiledPhase, 0, len(ps))
	for _, p := range ps {
		start, err := parseTimeOfDay(p.Start)
		if err != nil {
			return nil, fmt.Errorf("tradinghour: %s: start: %w", srcPath, err)
		}
		end, err := parseTimeOfDay(p.End)
		if err != nil {
			return nil, fmt.Errorf("tradinghour: %s: end: %w", srcPath, err)
		}
		out = append(out, compiledPhase{
			Session: Session(p.Session),
			Start:   start,
			End:     end,
		})
	}
	return out, nil
}

func loadHolidays() error {
	for mtype, mkt := range registry {
		dir := "data/holidays/" + marketDataDir(mtype)
		entries, err := fs.ReadDir(dataFS, dir)
		if err != nil {
			// Market with no holiday files is allowed (empty calendar).
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			p := path.Join(dir, e.Name())
			if err := loadHolidayFile(mkt, p); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadHolidayFile(mkt *Market, p string) error {
	raw, err := dataFS.ReadFile(p)
	if err != nil {
		return fmt.Errorf("tradinghour: read %s: %w", p, err)
	}
	var hf holidayFile
	if err := yaml.Unmarshal(raw, &hf); err != nil {
		return fmt.Errorf("tradinghour: parse %s: %w", p, err)
	}
	if MarketType(hf.Market) != mkt.Type {
		return fmt.Errorf("tradinghour: %s: market mismatch (%q vs %q)", p, hf.Market, mkt.Type)
	}
	for _, h := range hf.Holidays {
		t, err := time.Parse("2006-01-02", h.Date)
		if err != nil {
			return fmt.Errorf("tradinghour: %s: bad date %q: %w", p, h.Date, err)
		}
		cd := civilDate{year: t.Year(), month: int(t.Month()), day: t.Day()}
		mkt.Holidays[cd] = holidayEntry{
			Name: h.Name,
			Type: HolidayType(h.Type),
		}
	}
	return nil
}

// marketDataDir maps MarketType to the directory name under data/holidays/.
func marketDataDir(m MarketType) string {
	switch m {
	case MarketNASDAQ:
		return "nasdaq"
	case MarketHKEX:
		return "hkex"
	case MarketChinaAShare:
		return "china-ashare"
	case MarketTSE:
		return "tse"
	case MarketKRX:
		return "krx"
	default:
		return strings.ToLower(string(m))
	}
}
```

- [ ] **Step 3: Add loader test**

Create `loader_test.go`:
```go
package tradinghour

import "testing"

func TestRegistryLoaded(t *testing.T) {
	m, err := lookup(MarketNASDAQ)
	if err != nil {
		t.Fatalf("NASDAQ not registered: %v", err)
	}
	if m.Location.String() != "America/New_York" {
		t.Errorf("location = %v", m.Location)
	}
	// Monday should have 4 phases, Friday 3, Saturday 0, Sunday 1.
	if got := len(m.WeeklyPhases[int(0)]); got != 1 { // Sunday
		t.Errorf("Sunday phases = %d, want 1", got)
	}
	if got := len(m.WeeklyPhases[int(1)]); got != 4 { // Monday
		t.Errorf("Monday phases = %d, want 4", got)
	}
	if got := len(m.WeeklyPhases[int(5)]); got != 3 { // Friday
		t.Errorf("Friday phases = %d, want 3", got)
	}
	if got := len(m.WeeklyPhases[int(6)]); got != 0 { // Saturday
		t.Errorf("Saturday phases = %d, want 0", got)
	}
	if len(m.HalfDayPhases) != 2 {
		t.Errorf("HalfDayPhases len = %d, want 2", len(m.HalfDayPhases))
	}
}

func TestHolidaysLoaded(t *testing.T) {
	m, _ := lookup(MarketNASDAQ)
	h, ok := m.Holidays[civilDate{2026, 12, 25}]
	if !ok {
		t.Fatal("Christmas 2026 missing")
	}
	if h.Type != HolidayClosed {
		t.Errorf("Christmas type = %v", h.Type)
	}
	half, ok := m.Holidays[civilDate{2026, 11, 27}]
	if !ok {
		t.Fatal("Black Friday 2026 missing")
	}
	if half.Type != HolidayHalfDay {
		t.Errorf("Black Friday type = %v", half.Type)
	}
}

func TestLookupUnknown(t *testing.T) {
	if _, err := lookup(MarketType("NOPE")); err != ErrUnknownMarket {
		t.Errorf("err = %v, want ErrUnknownMarket", err)
	}
}
```

- [ ] **Step 4: Run tests — expect pass**

Run: `go test -run 'TestRegistryLoaded|TestHolidaysLoaded|TestLookupUnknown' ./...`
Expected: `PASS` (only NASDAQ is loaded at this point, which is fine).

- [ ] **Step 5: Commit**

```bash
git add market.go loader.go loader_test.go
git commit -m "feat: load embedded market schedules and holidays at init"
```

---

## Task 8: `Market.materialize` — weekday + holiday resolution

**Files:**
- Modify: `market.go`
- Create: `market_test.go`

- [ ] **Step 1: Write failing tests**

Create `market_test.go`:
```go
package tradinghour

import (
	"testing"
	"time"
)

func nasdaq(t *testing.T) *Market {
	t.Helper()
	m, err := lookup(MarketNASDAQ)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMaterializeRegularMonday(t *testing.T) {
	m := nasdaq(t)
	// 2026-03-02 is a Monday, no holidays.
	date := time.Date(2026, 3, 2, 0, 0, 0, 0, m.Location)
	phases, isHoliday, isHalfDay, name := m.materialize(date)
	if isHoliday || isHalfDay {
		t.Errorf("flags got (%v, %v), want (false, false)", isHoliday, isHalfDay)
	}
	if name != "" {
		t.Errorf("name = %q", name)
	}
	if len(phases) != 4 {
		t.Fatalf("phases len = %d, want 4", len(phases))
	}
	if phases[0].Session != SessionPreMarket {
		t.Errorf("phases[0].Session = %v", phases[0].Session)
	}
	if phases[3].Session != SessionOvernight {
		t.Errorf("phases[3].Session = %v", phases[3].Session)
	}
	// Overnight ends next calendar day.
	if phases[3].End.Day() != 3 {
		t.Errorf("overnight End.Day = %d, want 3", phases[3].End.Day())
	}
}

func TestMaterializeHolidayClosed(t *testing.T) {
	m := nasdaq(t)
	date := time.Date(2026, 12, 25, 0, 0, 0, 0, m.Location)
	phases, isHoliday, isHalfDay, name := m.materialize(date)
	if !isHoliday || isHalfDay {
		t.Errorf("flags got (%v, %v), want (true, false)", isHoliday, isHalfDay)
	}
	if name == "" {
		t.Error("holiday name empty")
	}
	if len(phases) != 0 {
		t.Errorf("phases len = %d, want 0", len(phases))
	}
}

func TestMaterializeHalfDay(t *testing.T) {
	m := nasdaq(t)
	// 2026-11-27 is Black Friday half-day.
	date := time.Date(2026, 11, 27, 0, 0, 0, 0, m.Location)
	phases, isHoliday, isHalfDay, name := m.materialize(date)
	if isHoliday || !isHalfDay {
		t.Errorf("flags got (%v, %v), want (false, true)", isHoliday, isHalfDay)
	}
	if name == "" {
		t.Error("name empty")
	}
	if len(phases) != 2 {
		t.Fatalf("phases len = %d, want 2", len(phases))
	}
	if phases[0].Session != SessionRegular ||
		phases[0].End.Hour() != 13 {
		t.Errorf("phases[0] = %+v", phases[0])
	}
}
```

- [ ] **Step 2: Run tests — expect failure**

Run: `go test -run TestMaterialize ./...`
Expected: compile error (method not defined).

- [ ] **Step 3: Implement `materialize` in `market.go`**

Append to `market.go`:
```go
// materialize returns the phases for `date` in this market's local timezone,
// plus holiday flags. The input date's Y/M/D is used; its hour/min are ignored.
func (m *Market) materialize(date time.Time) (phases []Phase, isHoliday bool, isHalfDay bool, name string) {
	y, mo, d := date.Date()
	localMidnight := time.Date(y, mo, d, 0, 0, 0, 0, m.Location)

	cd := civilDate{year: y, month: int(mo), day: d}
	if h, ok := m.Holidays[cd]; ok {
		switch h.Type {
		case HolidayClosed:
			return nil, true, false, h.Name
		case HolidayHalfDay:
			return instantiateAll(m.HalfDayPhases, localMidnight, m.Location), false, true, h.Name
		}
	}

	base := m.WeeklyPhases[int(localMidnight.Weekday())]
	return instantiateAll(base, localMidnight, m.Location), false, false, ""
}

func instantiateAll(cps []compiledPhase, date time.Time, loc *time.Location) []Phase {
	out := make([]Phase, len(cps))
	for i, c := range cps {
		out[i] = c.instantiate(date, loc)
	}
	return out
}
```

- [ ] **Step 4: Run tests — expect pass**

Run: `go test -run TestMaterialize ./...`
Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add market.go market_test.go
git commit -m "feat: add Market.materialize with weekly + holiday resolution"
```

---

## Task 9: `IsOpen` implementation

**Files:**
- Modify: `tradinghour.go`
- Create: `isopen_test.go`

- [ ] **Step 1: Write failing tests**

Create `isopen_test.go`:
```go
package tradinghour

import (
	"testing"
	"time"
)

func mustNY(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestIsOpenNASDAQ(t *testing.T) {
	ny := mustNY(t)
	cases := []struct {
		name      string
		local     time.Time
		wantOpen  bool
		wantSess  Session
	}{
		{"Mon 03:59 closed", time.Date(2026, 3, 2, 3, 59, 0, 0, ny), false, SessionClosed},
		{"Mon 04:00 premarket start", time.Date(2026, 3, 2, 4, 0, 0, 0, ny), true, SessionPreMarket},
		{"Mon 09:29 premarket end", time.Date(2026, 3, 2, 9, 29, 59, 0, ny), true, SessionPreMarket},
		{"Mon 09:30 regular start", time.Date(2026, 3, 2, 9, 30, 0, 0, ny), true, SessionRegular},
		{"Mon 12:00 regular", time.Date(2026, 3, 2, 12, 0, 0, 0, ny), true, SessionRegular},
		{"Mon 16:00 postmarket start", time.Date(2026, 3, 2, 16, 0, 0, 0, ny), true, SessionPostMarket},
		{"Mon 20:00 overnight start", time.Date(2026, 3, 2, 20, 0, 0, 0, ny), true, SessionOvernight},
		{"Tue 02:00 overnight (spillover)", time.Date(2026, 3, 3, 2, 0, 0, 0, ny), true, SessionOvernight},
		{"Tue 04:00 closed", time.Date(2026, 3, 3, 4, 0, 0, 0, ny), false, SessionClosed},
		{"Sat any time closed", time.Date(2026, 3, 7, 10, 0, 0, 0, ny), false, SessionClosed},
		{"Fri 21:00 NO overnight (BOAT Sun-Thu)", time.Date(2026, 3, 6, 21, 0, 0, 0, ny), false, SessionClosed},
		{"Sun 21:00 overnight", time.Date(2026, 3, 8, 21, 0, 0, 0, ny), true, SessionOvernight},
		{"Christmas Day closed", time.Date(2026, 12, 25, 11, 0, 0, 0, ny), false, SessionClosed},
		{"Black Friday 12:59 regular (half-day)", time.Date(2026, 11, 27, 12, 59, 0, 0, ny), true, SessionRegular},
		{"Black Friday 13:00 postmarket (half-day)", time.Date(2026, 11, 27, 13, 0, 0, 0, ny), true, SessionPostMarket},
		{"Black Friday 17:00 closed (half-day)", time.Date(2026, 11, 27, 17, 0, 0, 0, ny), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.local.Unix(), MarketNASDAQ)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
			if st.Market != MarketNASDAQ {
				t.Errorf("Market = %v", st.Market)
			}
		})
	}
}

func TestIsOpenUnknownMarket(t *testing.T) {
	_, err := IsOpen(0, MarketType("BOGUS"))
	if err != ErrUnknownMarket {
		t.Errorf("err = %v", err)
	}
}
```

- [ ] **Step 2: Run tests — expect fail**

Run: `go test -run TestIsOpen ./...`
Expected: panic ("not implemented") or fail.

- [ ] **Step 3: Implement `IsOpen` in `tradinghour.go`**

Replace the placeholder `IsOpen` function with:
```go
// IsOpen reports the market's status at the given absolute instant.
func IsOpen(unixSec int64, m MarketType) (Status, error) {
	mkt, err := lookup(m)
	if err != nil {
		return Status{}, err
	}
	t := time.Unix(unixSec, 0).In(mkt.Location)
	today, _, _, _ := mkt.materialize(t)
	yesterday, _, _, _ := mkt.materialize(t.AddDate(0, 0, -1))
	for _, p := range append(yesterday, today...) {
		if (t.Equal(p.Start) || t.After(p.Start)) && t.Before(p.End) {
			return Status{Open: true, Session: p.Session, Market: m}, nil
		}
	}
	return Status{Open: false, Session: SessionClosed, Market: m}, nil
}
```

- [ ] **Step 4: Run tests — expect pass**

Run: `go test -run TestIsOpen ./...`
Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add tradinghour.go isopen_test.go
git commit -m "feat: implement IsOpen with overnight spillover handling"
```

---

## Task 10: `Timeline` implementation

**Files:**
- Modify: `tradinghour.go`
- Create: `timeline_test.go`

- [ ] **Step 1: Write failing tests**

Create `timeline_test.go`:
```go
package tradinghour

import (
	"testing"
	"time"
)

func TestTimelineRegularMonday(t *testing.T) {
	ny := mustNY(t)
	ds, err := Timeline(time.Date(2026, 3, 2, 15, 17, 0, 0, ny), MarketNASDAQ) // hour/min ignored
	if err != nil {
		t.Fatal(err)
	}
	if ds.Date.Hour() != 0 || ds.Date.Minute() != 0 {
		t.Errorf("Date not midnight: %v", ds.Date)
	}
	if ds.Market != MarketNASDAQ {
		t.Errorf("Market = %v", ds.Market)
	}
	if ds.IsHoliday || ds.IsHalfDay {
		t.Errorf("flags = (%v, %v)", ds.IsHoliday, ds.IsHalfDay)
	}
	if len(ds.Phases) != 4 {
		t.Fatalf("phases = %d", len(ds.Phases))
	}
	if ds.Phases[3].Session != SessionOvernight || ds.Phases[3].End.Day() != 3 {
		t.Errorf("overnight phase = %+v", ds.Phases[3])
	}
}

func TestTimelineChristmas(t *testing.T) {
	ny := mustNY(t)
	ds, _ := Timeline(time.Date(2026, 12, 25, 0, 0, 0, 0, ny), MarketNASDAQ)
	if !ds.IsHoliday || ds.IsHalfDay {
		t.Errorf("flags = (%v, %v)", ds.IsHoliday, ds.IsHalfDay)
	}
	if len(ds.Phases) != 0 {
		t.Errorf("phases = %d", len(ds.Phases))
	}
	if ds.HolidayName == "" {
		t.Error("name empty")
	}
}

func TestTimelineBlackFridayHalfDay(t *testing.T) {
	ny := mustNY(t)
	ds, _ := Timeline(time.Date(2026, 11, 27, 0, 0, 0, 0, ny), MarketNASDAQ)
	if ds.IsHoliday || !ds.IsHalfDay {
		t.Errorf("flags = (%v, %v)", ds.IsHoliday, ds.IsHalfDay)
	}
	if len(ds.Phases) != 2 {
		t.Fatalf("phases = %d", len(ds.Phases))
	}
	if ds.Phases[1].End.Hour() != 17 {
		t.Errorf("postmarket end = %v", ds.Phases[1].End)
	}
}

func TestTimelineIgnoresInputTZ(t *testing.T) {
	// Passing a UTC time whose Y/M/D differs from NY's Y/M/D should still resolve
	// using the Y/M/D values from the input (not converted into market tz).
	// 2026-03-02 02:00 UTC == 2026-03-01 21:00 NY, but our rule is "input's Y/M/D
	// is the date". So Timeline should treat this as March 2 in NY.
	utc := time.Date(2026, 3, 2, 2, 0, 0, 0, time.UTC)
	ds, _ := Timeline(utc, MarketNASDAQ)
	if ds.Date.Month() != 3 || ds.Date.Day() != 2 {
		t.Errorf("Date = %v, want 2026-03-02", ds.Date)
	}
}
```

- [ ] **Step 2: Run tests — expect fail**

Run: `go test -run TestTimeline ./...`
Expected: panic.

- [ ] **Step 3: Implement `Timeline`**

Replace placeholder in `tradinghour.go`:
```go
// Timeline returns the full day schedule for the given date, interpreted in the
// market's local timezone (only Y/M/D of date is used; hour/min/sec/loc are ignored).
func Timeline(date time.Time, m MarketType) (DaySchedule, error) {
	mkt, err := lookup(m)
	if err != nil {
		return DaySchedule{}, err
	}
	y, mo, d := date.Date()
	localMidnight := time.Date(y, mo, d, 0, 0, 0, 0, mkt.Location)
	phases, isHoliday, isHalfDay, name := mkt.materialize(localMidnight)
	return DaySchedule{
		Date:        localMidnight,
		Market:      m,
		Phases:      phases,
		IsHoliday:   isHoliday,
		IsHalfDay:   isHalfDay,
		HolidayName: name,
	}, nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test -run TestTimeline ./...`
Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add tradinghour.go timeline_test.go
git commit -m "feat: implement Timeline returning DaySchedule"
```

---

## Task 11: `NextOpen` and `NextClose`

**Files:**
- Modify: `tradinghour.go`
- Create: `nextopen_test.go`

- [ ] **Step 1: Write failing tests**

Create `nextopen_test.go`:
```go
package tradinghour

import (
	"testing"
	"time"
)

func TestNextOpenWhenClosed(t *testing.T) {
	ny := mustNY(t)
	// Sat 10:00 NY - next open is Sun 20:00 (overnight).
	got, err := NextOpen(time.Date(2026, 3, 7, 10, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 3, 8, 20, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextOpenWhenOpen(t *testing.T) {
	ny := mustNY(t)
	// During regular hours, NextOpen is the start of the *next* phase after current.
	// Mon 11:00 is in regular (09:30-16:00). Next open phase is postmarket 16:00.
	got, _ := NextOpen(time.Date(2026, 3, 2, 11, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	want := time.Date(2026, 3, 2, 16, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextOpenSkipsHoliday(t *testing.T) {
	ny := mustNY(t)
	// Christmas 2026 is Fri. NextOpen called on Christmas should skip the closed
	// day and land on Sunday 20:00 overnight (Dec 27 20:00).
	got, _ := NextOpen(time.Date(2026, 12, 25, 10, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	want := time.Date(2026, 12, 27, 20, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextCloseWhenOpen(t *testing.T) {
	ny := mustNY(t)
	// Mon 11:00 - inside regular (09:30-16:00). NextClose = 16:00.
	got, _ := NextClose(time.Date(2026, 3, 2, 11, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	want := time.Date(2026, 3, 2, 16, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextCloseWhenClosed(t *testing.T) {
	ny := mustNY(t)
	// Sat 10:00 - market closed. NextClose = end of the next open phase
	// (Sun 20:00 overnight ends Mon 04:00).
	got, _ := NextClose(time.Date(2026, 3, 7, 10, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	want := time.Date(2026, 3, 9, 4, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
```

- [ ] **Step 2: Run tests — expect fail**

Run: `go test -run 'TestNextOpen|TestNextClose' ./...`
Expected: panic.

- [ ] **Step 3: Implement NextOpen / NextClose**

Replace placeholders in `tradinghour.go`:
```go
const searchHorizonDays = 15

// NextOpen returns the nearest future instant at which IsOpen transitions to true.
func NextOpen(unixSec int64, m MarketType) (time.Time, error) {
	mkt, err := lookup(m)
	if err != nil {
		return time.Time{}, err
	}
	t := time.Unix(unixSec, 0).In(mkt.Location)
	for i := 0; i < searchHorizonDays; i++ {
		day := time.Date(t.Year(), t.Month(), t.Day()+i, 0, 0, 0, 0, mkt.Location)
		phases, _, _, _ := mkt.materialize(day)
		for _, p := range phases {
			if p.Start.After(t) {
				return p.Start, nil
			}
		}
	}
	return time.Time{}, nil // no open phase in horizon; MVP treats as empty result
}

// NextClose returns the end of the current open phase, or the end of the next open phase if closed.
func NextClose(unixSec int64, m MarketType) (time.Time, error) {
	mkt, err := lookup(m)
	if err != nil {
		return time.Time{}, err
	}
	t := time.Unix(unixSec, 0).In(mkt.Location)
	for i := -1; i < searchHorizonDays; i++ {
		day := time.Date(t.Year(), t.Month(), t.Day()+i, 0, 0, 0, 0, mkt.Location)
		phases, _, _, _ := mkt.materialize(day)
		for _, p := range phases {
			if p.End.After(t) {
				return p.End, nil
			}
		}
	}
	return time.Time{}, nil
}
```

Note: `NextClose` starts iterating at `i = -1` so that an overnight phase starting yesterday is considered when called during the early-morning overnight window.

- [ ] **Step 4: Run tests**

Run: `go test -run 'TestNextOpen|TestNextClose' ./...`
Expected: `PASS`.

- [ ] **Step 5: Run full test suite**

Run: `go test ./...`
Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add tradinghour.go nextopen_test.go
git commit -m "feat: implement NextOpen and NextClose"
```

---

## Task 12: HKEX data + tests

**Files:**
- Create: `data/markets/hkex.yaml`, `data/holidays/hkex/2026.yaml`, `hkex_test.go`

- [ ] **Step 1: Create `data/markets/hkex.yaml`**

HKEX equity hours (Mon-Fri): 09:30-12:00 morning, 13:00-16:10 afternoon (includes closing auction 16:00-16:10). Lunch closed. Half-day schedule (LNY Eve, Xmas Eve) is morning only closing 12:00.

```yaml
market: HKEX
timezone: Asia/Hong_Kong
weekly_schedule:
  monday: &hk_full
    - {session: regular, start: "09:30", end: "12:00"}
    - {session: regular, start: "13:00", end: "16:10"}
  tuesday: *hk_full
  wednesday: *hk_full
  thursday: *hk_full
  friday: *hk_full
  saturday: []
  sunday: []
half_day_schedule:
  - {session: regular, start: "09:30", end: "12:00"}
```

- [ ] **Step 2: Create `data/holidays/hkex/2026.yaml`**

Based on [HKEX 2026 calendar](https://www.hkex.com.hk/Services/Trading-hours-and-Severe-Weather-Arrangements/Trading-Hours/Holiday-Schedule?sc_lang=en):
```yaml
market: HKEX
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",           type: closed}
  - {date: "2026-02-17", name: "Lunar New Year Day 1",     type: closed}
  - {date: "2026-02-18", name: "Lunar New Year Day 2",     type: closed}
  - {date: "2026-02-19", name: "Lunar New Year Day 3",     type: closed}
  - {date: "2026-04-03", name: "Good Friday",              type: closed}
  - {date: "2026-04-06", name: "Easter Monday",            type: closed}
  - {date: "2026-04-07", name: "Ching Ming Festival (obs.)", type: closed}
  - {date: "2026-05-01", name: "Labour Day",               type: closed}
  - {date: "2026-05-25", name: "Buddha's Birthday",        type: closed}
  - {date: "2026-06-19", name: "Tuen Ng Festival",         type: closed}
  - {date: "2026-07-01", name: "HKSAR Establishment Day",  type: closed}
  - {date: "2026-09-25", name: "Day after Mid-Autumn",     type: closed}
  - {date: "2026-10-01", name: "National Day",             type: closed}
  - {date: "2026-10-19", name: "Chung Yeung Festival",     type: closed}
  - {date: "2026-12-24", name: "Christmas Eve",            type: half_day}
  - {date: "2026-12-25", name: "Christmas Day",            type: closed}
  - {date: "2026-12-26", name: "Boxing Day (obs.)",        type: closed}
```

(The PR-reviewing human should cross-check with HKEX's official 2026 schedule when published.)

- [ ] **Step 3: Create `hkex_test.go`**

```go
package tradinghour

import (
	"testing"
	"time"
)

func mustHK(t *testing.T) *time.Location {
	t.Helper()
	loc, _ := time.LoadLocation("Asia/Hong_Kong")
	return loc
}

func TestIsOpenHKEX(t *testing.T) {
	hk := mustHK(t)
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
		wantSess Session
	}{
		{"Mon 09:29 closed", time.Date(2026, 3, 2, 9, 29, 0, 0, hk), false, SessionClosed},
		{"Mon 09:30 open", time.Date(2026, 3, 2, 9, 30, 0, 0, hk), true, SessionRegular},
		{"Mon 12:00 lunch", time.Date(2026, 3, 2, 12, 0, 0, 0, hk), false, SessionClosed},
		{"Mon 12:59 lunch", time.Date(2026, 3, 2, 12, 59, 0, 0, hk), false, SessionClosed},
		{"Mon 13:00 open", time.Date(2026, 3, 2, 13, 0, 0, 0, hk), true, SessionRegular},
		{"Mon 16:09 open (auction)", time.Date(2026, 3, 2, 16, 9, 59, 0, hk), true, SessionRegular},
		{"Mon 16:10 closed", time.Date(2026, 3, 2, 16, 10, 0, 0, hk), false, SessionClosed},
		{"LNY day 1 closed", time.Date(2026, 2, 17, 10, 0, 0, 0, hk), false, SessionClosed},
		{"Xmas Eve 13:00 closed (half-day)", time.Date(2026, 12, 24, 13, 0, 0, 0, hk), false, SessionClosed},
		{"Xmas Eve 11:59 open (half-day)", time.Date(2026, 12, 24, 11, 59, 0, 0, hk), true, SessionRegular},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, _ := IsOpen(c.local.Unix(), MarketHKEX)
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("got (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test -run TestIsOpenHKEX ./...`
Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add data/markets/hkex.yaml data/holidays/hkex/2026.yaml hkex_test.go
git commit -m "data: add HKEX schedule, 2026 holidays, and tests"
```

---

## Task 13: China A-Share data + tests

**Files:**
- Create: `data/markets/china-ashare.yaml`, `data/holidays/china-ashare/2026.yaml`, `china_ashare_test.go`

- [ ] **Step 1: Create `data/markets/china-ashare.yaml`**

```yaml
market: ChinaAShare
timezone: Asia/Shanghai
weekly_schedule:
  monday: &cn_full
    - {session: regular, start: "09:30", end: "11:30"}
    - {session: regular, start: "13:00", end: "15:00"}
  tuesday: *cn_full
  wednesday: *cn_full
  thursday: *cn_full
  friday: *cn_full
  saturday: []
  sunday: []
```

No `half_day_schedule` — A-Shares don't use half-days.

- [ ] **Step 2: Create `data/holidays/china-ashare/2026.yaml`**

2026 SSE/SZSE holidays (based on published calendar):
```yaml
market: ChinaAShare
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",            type: closed}
  - {date: "2026-02-16", name: "Spring Festival Eve",       type: closed}
  - {date: "2026-02-17", name: "Spring Festival",           type: closed}
  - {date: "2026-02-18", name: "Spring Festival",           type: closed}
  - {date: "2026-02-19", name: "Spring Festival",           type: closed}
  - {date: "2026-02-20", name: "Spring Festival",           type: closed}
  - {date: "2026-02-23", name: "Spring Festival (obs.)",    type: closed}
  - {date: "2026-02-24", name: "Spring Festival (obs.)",    type: closed}
  - {date: "2026-04-06", name: "Qingming Festival (obs.)",  type: closed}
  - {date: "2026-05-01", name: "Labour Day",                type: closed}
  - {date: "2026-05-04", name: "Labour Day (obs.)",         type: closed}
  - {date: "2026-05-05", name: "Labour Day (obs.)",         type: closed}
  - {date: "2026-06-19", name: "Dragon Boat Festival",      type: closed}
  - {date: "2026-09-25", name: "Mid-Autumn (obs.)",         type: closed}
  - {date: "2026-10-01", name: "National Day",              type: closed}
  - {date: "2026-10-02", name: "National Day",              type: closed}
  - {date: "2026-10-05", name: "National Day Break",        type: closed}
  - {date: "2026-10-06", name: "National Day Break",        type: closed}
  - {date: "2026-10-07", name: "National Day Break",        type: closed}
  - {date: "2026-10-08", name: "National Day Break",        type: closed}
```

Note for reviewer: Chinese holidays frequently have "make-up workdays" on weekends — those are still trading days and need no entry. The human reviewing the GH-Action PR must cross-check against the official State Council notice.

- [ ] **Step 3: Create `china_ashare_test.go`**

```go
package tradinghour

import (
	"testing"
	"time"
)

func mustSH(t *testing.T) *time.Location {
	t.Helper()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return loc
}

func TestIsOpenChinaAShare(t *testing.T) {
	sh := mustSH(t)
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
		wantSess Session
	}{
		{"Mon 09:29 closed", time.Date(2026, 3, 2, 9, 29, 0, 0, sh), false, SessionClosed},
		{"Mon 09:30 open", time.Date(2026, 3, 2, 9, 30, 0, 0, sh), true, SessionRegular},
		{"Mon 11:30 lunch", time.Date(2026, 3, 2, 11, 30, 0, 0, sh), false, SessionClosed},
		{"Mon 13:00 afternoon open", time.Date(2026, 3, 2, 13, 0, 0, 0, sh), true, SessionRegular},
		{"Mon 15:00 closed", time.Date(2026, 3, 2, 15, 0, 0, 0, sh), false, SessionClosed},
		{"Spring Festival closed", time.Date(2026, 2, 17, 10, 0, 0, 0, sh), false, SessionClosed},
		{"Golden Week Oct 5", time.Date(2026, 10, 5, 10, 0, 0, 0, sh), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, _ := IsOpen(c.local.Unix(), MarketChinaAShare)
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("got (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test -run TestIsOpenChinaAShare ./...`
Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add data/markets/china-ashare.yaml data/holidays/china-ashare/2026.yaml china_ashare_test.go
git commit -m "data: add China A-Share schedule, 2026 holidays, and tests"
```

---

## Task 14: TSE (Tokyo) data + tests

**Files:**
- Create: `data/markets/tse.yaml`, `data/holidays/tse/2026.yaml`, `tse_test.go`

- [ ] **Step 1: Create `data/markets/tse.yaml`**

```yaml
market: TSE
timezone: Asia/Tokyo
weekly_schedule:
  monday: &jp_full
    - {session: regular, start: "09:00", end: "11:30"}
    - {session: regular, start: "12:30", end: "15:30"}
  tuesday: *jp_full
  wednesday: *jp_full
  thursday: *jp_full
  friday: *jp_full
  saturday: []
  sunday: []
```

- [ ] **Step 2: Create `data/holidays/tse/2026.yaml`**

Based on the Japan "National Holidays Act" dates for 2026:
```yaml
market: TSE
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",     type: closed}
  - {date: "2026-01-02", name: "Bank holiday",       type: closed}
  - {date: "2026-01-12", name: "Coming of Age Day",  type: closed}
  - {date: "2026-02-11", name: "Foundation Day",     type: closed}
  - {date: "2026-02-23", name: "Emperor's Birthday", type: closed}
  - {date: "2026-03-20", name: "Vernal Equinox",     type: closed}
  - {date: "2026-04-29", name: "Showa Day",          type: closed}
  - {date: "2026-05-04", name: "Greenery Day",       type: closed}
  - {date: "2026-05-05", name: "Children's Day",     type: closed}
  - {date: "2026-05-06", name: "Constitution Day (obs.)", type: closed}
  - {date: "2026-07-20", name: "Marine Day",         type: closed}
  - {date: "2026-08-11", name: "Mountain Day",       type: closed}
  - {date: "2026-09-21", name: "Respect for the Aged Day", type: closed}
  - {date: "2026-09-22", name: "Autumnal Equinox",   type: closed}
  - {date: "2026-10-12", name: "Health-Sports Day",  type: closed}
  - {date: "2026-11-03", name: "Culture Day",        type: closed}
  - {date: "2026-11-23", name: "Labour Thanksgiving", type: closed}
  - {date: "2026-12-31", name: "Year-end holiday",   type: closed}
```

- [ ] **Step 3: Create `tse_test.go`**

```go
package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenTSE(t *testing.T) {
	jp, _ := time.LoadLocation("Asia/Tokyo")
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
	}{
		{"Mon 08:59 closed", time.Date(2026, 3, 2, 8, 59, 0, 0, jp), false},
		{"Mon 09:00 open", time.Date(2026, 3, 2, 9, 0, 0, 0, jp), true},
		{"Mon 11:30 lunch", time.Date(2026, 3, 2, 11, 30, 0, 0, jp), false},
		{"Mon 12:30 open", time.Date(2026, 3, 2, 12, 30, 0, 0, jp), true},
		{"Mon 15:30 closed", time.Date(2026, 3, 2, 15, 30, 0, 0, jp), false},
		{"Golden Week May 5", time.Date(2026, 5, 5, 10, 0, 0, 0, jp), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, _ := IsOpen(c.local.Unix(), MarketTSE)
			if st.Open != c.wantOpen {
				t.Errorf("got open=%v, want %v", st.Open, c.wantOpen)
			}
		})
	}
}
```

- [ ] **Step 4: Run + commit**

Run: `go test -run TestIsOpenTSE ./...`  → Expected: `PASS`.

```bash
git add data/markets/tse.yaml data/holidays/tse/2026.yaml tse_test.go
git commit -m "data: add TSE schedule, 2026 holidays, and tests"
```

---

## Task 15: KRX data + tests

**Files:**
- Create: `data/markets/krx.yaml`, `data/holidays/krx/2026.yaml`, `krx_test.go`

- [ ] **Step 1: Create `data/markets/krx.yaml`**

```yaml
market: KRX
timezone: Asia/Seoul
weekly_schedule:
  monday: &kr_full
    - {session: premarket,  start: "08:00", end: "09:00"}
    - {session: regular,    start: "09:00", end: "15:30"}
    - {session: postmarket, start: "15:40", end: "18:00"}
  tuesday: *kr_full
  wednesday: *kr_full
  thursday: *kr_full
  friday: *kr_full
  saturday: []
  sunday: []
```

- [ ] **Step 2: Create `data/holidays/krx/2026.yaml`**

```yaml
market: KRX
year: 2026
holidays:
  - {date: "2026-01-01", name: "New Year's Day",         type: closed}
  - {date: "2026-02-16", name: "Lunar New Year (obs.)",  type: closed}
  - {date: "2026-02-17", name: "Lunar New Year",         type: closed}
  - {date: "2026-02-18", name: "Lunar New Year",         type: closed}
  - {date: "2026-03-01", name: "Independence Movement",  type: closed}
  - {date: "2026-05-01", name: "Labour Day",             type: closed}
  - {date: "2026-05-05", name: "Children's Day",         type: closed}
  - {date: "2026-05-25", name: "Buddha's Birthday (obs.)", type: closed}
  - {date: "2026-06-03", name: "Election Day",           type: closed}
  - {date: "2026-06-06", name: "Memorial Day",           type: closed}
  - {date: "2026-08-15", name: "Liberation Day",         type: closed}
  - {date: "2026-09-24", name: "Chuseok (obs.)",         type: closed}
  - {date: "2026-09-25", name: "Chuseok",                type: closed}
  - {date: "2026-09-26", name: "Chuseok",                type: closed}
  - {date: "2026-10-03", name: "National Foundation",    type: closed}
  - {date: "2026-10-09", name: "Hangul Day",             type: closed}
  - {date: "2026-12-25", name: "Christmas Day",          type: closed}
  - {date: "2026-12-31", name: "Year-end holiday",       type: closed}
```

- [ ] **Step 3: Create `krx_test.go`**

```go
package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenKRX(t *testing.T) {
	kr, _ := time.LoadLocation("Asia/Seoul")
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
		wantSess Session
	}{
		{"Mon 07:59 closed", time.Date(2026, 3, 2, 7, 59, 0, 0, kr), false, SessionClosed},
		{"Mon 08:00 pre", time.Date(2026, 3, 2, 8, 0, 0, 0, kr), true, SessionPreMarket},
		{"Mon 09:00 regular", time.Date(2026, 3, 2, 9, 0, 0, 0, kr), true, SessionRegular},
		{"Mon 15:30 closed", time.Date(2026, 3, 2, 15, 30, 0, 0, kr), false, SessionClosed},
		{"Mon 15:35 gap closed", time.Date(2026, 3, 2, 15, 35, 0, 0, kr), false, SessionClosed},
		{"Mon 15:40 post", time.Date(2026, 3, 2, 15, 40, 0, 0, kr), true, SessionPostMarket},
		{"Mon 18:00 closed", time.Date(2026, 3, 2, 18, 0, 0, 0, kr), false, SessionClosed},
		{"Chuseok closed", time.Date(2026, 9, 25, 10, 0, 0, 0, kr), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, _ := IsOpen(c.local.Unix(), MarketKRX)
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("got (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
```

- [ ] **Step 4: Run + commit**

Run: `go test -run TestIsOpenKRX ./...`  → `PASS`.

```bash
git add data/markets/krx.yaml data/holidays/krx/2026.yaml krx_test.go
git commit -m "data: add KRX schedule, 2026 holidays, and tests"
```

---

## Task 16: Cross-market integration tests (DST, round-trip consistency)

**Files:**
- Create: `integration_test.go`

- [ ] **Step 1: Write tests**

```go
package tradinghour

import (
	"testing"
	"time"
)

// Round-trip: for every phase in a day's timeline, IsOpen at Start should
// return open with the phase's session, and IsOpen at End should return closed
// (or open-on-next-phase if phases are contiguous).
func TestTimelineIsOpenConsistency(t *testing.T) {
	markets := []MarketType{MarketNASDAQ, MarketHKEX, MarketChinaAShare, MarketTSE, MarketKRX}
	// Pick a normal trading Monday (2026-03-02) — no holidays any market.
	dates := []time.Time{
		time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
	}
	for _, m := range markets {
		for _, d := range dates {
			ds, err := Timeline(d, m)
			if err != nil {
				t.Fatalf("%s Timeline: %v", m, err)
			}
			for i, p := range ds.Phases {
				st, _ := IsOpen(p.Start.Unix(), m)
				if !st.Open || st.Session != p.Session {
					t.Errorf("%s phase[%d] start: got (%v, %v), want (true, %v)", m, i, st.Open, st.Session, p.Session)
				}
				// One nanosecond before End: still in this phase.
				st, _ = IsOpen(p.End.Add(-time.Nanosecond).Unix(), m)
				if !st.Open || st.Session != p.Session {
					t.Errorf("%s phase[%d] end-1ns: got (%v, %v), want (true, %v)", m, i, st.Open, st.Session, p.Session)
				}
			}
		}
	}
}

// DST spring-forward: on 2026-03-08, NASDAQ clocks jump from 02:00 -> 03:00 EST -> EDT.
// IsOpen at 04:00 Sun EDT (overnight runs through Mon 04:00) and Mon 09:30 EDT should
// correctly reflect the post-DST offsets.
func TestNASDAQSpringForward(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	mon := time.Date(2026, 3, 9, 9, 30, 0, 0, ny)
	st, _ := IsOpen(mon.Unix(), MarketNASDAQ)
	if !st.Open || st.Session != SessionRegular {
		t.Errorf("Mon 09:30 EDT: got (%v, %v), want (true, regular)", st.Open, st.Session)
	}
	// Sunday overnight ends Mon 04:00 EDT. At 03:59 EDT Mon, still overnight.
	at := time.Date(2026, 3, 9, 3, 59, 0, 0, ny)
	st, _ = IsOpen(at.Unix(), MarketNASDAQ)
	if !st.Open || st.Session != SessionOvernight {
		t.Errorf("Mon 03:59 EDT: got (%v, %v), want (true, overnight)", st.Open, st.Session)
	}
}

// DST fall-back: on 2026-11-01, NASDAQ clocks shift 02:00 EDT -> 01:00 EST.
func TestNASDAQFallBack(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	mon := time.Date(2026, 11, 2, 9, 30, 0, 0, ny)
	st, _ := IsOpen(mon.Unix(), MarketNASDAQ)
	if !st.Open || st.Session != SessionRegular {
		t.Errorf("Mon 09:30 post-fallback: got (%v, %v)", st.Open, st.Session)
	}
}
```

- [ ] **Step 2: Run**

Run: `go test -run 'TestTimelineIsOpenConsistency|TestNASDAQSpringForward|TestNASDAQFallBack' ./...`
Expected: `PASS`.

- [ ] **Step 3: Run full suite with race detector**

Run: `go test -race ./...`
Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add integration_test.go
git commit -m "test: add DST and round-trip integration tests"
```

---

## Task 17: Holiday refresh script (`scripts/refresh_holidays.py`)

**Files:**
- Create: `scripts/refresh_holidays.py`

- [ ] **Step 1: Write the script**

```python
#!/usr/bin/env python3
"""Generate data/holidays/<market>/<year>.yaml from exchange_calendars.

Usage: python scripts/refresh_holidays.py --year 2027 [--out-dir data/holidays]

For each supported market, emits a YAML file listing closed days and
(where detectable) half-days. Half-days require human verification.
"""

import argparse
import datetime as dt
import os
from pathlib import Path

import exchange_calendars as ec
import yaml

# Map our MarketType -> (ec code, data dir)
MARKETS = {
    "NASDAQ":      ("XNAS", "nasdaq"),
    "HKEX":        ("XHKG", "hkex"),
    "ChinaAShare": ("XSHG", "china-ashare"),
    "TSE":         ("XTKS", "tse"),
    "KRX":         ("XKRX", "krx"),
}

# Known half-day markets and their fallback "early close" detection.
# exchange_calendars exposes .early_closes for NYSE/NASDAQ; for HKEX we flag
# manually below in MANUAL_HALF_DAYS (script reviewer verifies).
MANUAL_HALF_DAYS = {
    "HKEX": {
        # date -> name. Lunar New Year Eve and Christmas Eve are traditional.
        # The reviewer must update this set each year.
    },
}

def holidays_for(market: str, year: int) -> list[dict]:
    ec_code, _ = MARKETS[market]
    cal = ec.get_calendar(ec_code)
    start = dt.date(year, 1, 1)
    end = dt.date(year, 12, 31)
    all_dates = {d.date() for d in cal.schedule.loc[str(start):str(end)].index}
    calendar_days = {start + dt.timedelta(days=i) for i in range((end - start).days + 1)}
    closed_weekdays = {d for d in calendar_days
                       if d.weekday() < 5 and d not in all_dates}

    early = set()
    if hasattr(cal, "early_closes"):
        ec_early = cal.early_closes
        if ec_early is not None:
            early = {ts.date() for ts in ec_early
                     if start <= ts.date() <= end}
    for d in MANUAL_HALF_DAYS.get(market, {}):
        if start <= d <= end:
            early.add(d)

    entries = []
    for d in sorted(closed_weekdays):
        entries.append({
            "date": d.isoformat(),
            "name": holiday_name(cal, d),
            "type": "closed",
        })
    for d in sorted(early):
        # Override a closed entry if this date appeared there (unlikely).
        entries = [e for e in entries if e["date"] != d.isoformat()]
        entries.append({
            "date": d.isoformat(),
            "name": MANUAL_HALF_DAYS.get(market, {}).get(d) or "Early close",
            "type": "half_day",
        })
    entries.sort(key=lambda e: e["date"])
    return entries

def holiday_name(cal, date: dt.date) -> str:
    # exchange_calendars doesn't expose names directly; fall back to generic.
    return "Market holiday"

def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--year", type=int, required=True)
    ap.add_argument("--out-dir", type=Path, default=Path("data/holidays"))
    args = ap.parse_args()

    for market, (_, subdir) in MARKETS.items():
        out = args.out_dir / subdir / f"{args.year}.yaml"
        out.parent.mkdir(parents=True, exist_ok=True)
        payload = {
            "market": market,
            "year": args.year,
            "holidays": holidays_for(market, args.year),
        }
        with out.open("w") as f:
            yaml.safe_dump(payload, f, sort_keys=False, allow_unicode=True)
        print(f"wrote {out} ({len(payload['holidays'])} entries)")

if __name__ == "__main__":
    main()
```

- [ ] **Step 2: Smoke-test locally (optional, skippable if no Python env)**

Run:
```bash
python -m venv /tmp/th-venv
/tmp/th-venv/bin/pip install exchange_calendars pyyaml
/tmp/th-venv/bin/python scripts/refresh_holidays.py --year 2027 --out-dir /tmp/th-out
ls /tmp/th-out/
```

If Python is not available locally, skip this step — the GH Action will exercise the script.

- [ ] **Step 3: Commit**

```bash
git add scripts/refresh_holidays.py
git commit -m "feat: add Python script to generate holiday YAMLs from exchange_calendars"
```

---

## Task 18: GitHub Action workflow

**Files:**
- Create: `.github/workflows/refresh-holidays.yml`

- [ ] **Step 1: Write workflow**

```yaml
name: Refresh holiday calendars

on:
  schedule:
    - cron: '0 0 15 11 *'  # Nov 15 UTC every year
  workflow_dispatch:
    inputs:
      year:
        description: "Year to generate (default: next year)"
        required: false

jobs:
  refresh:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-python@v5
        with:
          python-version: "3.11"

      - name: Install deps
        run: pip install exchange_calendars pyyaml

      - name: Pick year
        id: year
        run: |
          if [ -n "${{ inputs.year }}" ]; then
            echo "year=${{ inputs.year }}" >> "$GITHUB_OUTPUT"
          else
            echo "year=$(date -u -d '+1 year' +%Y)" >> "$GITHUB_OUTPUT"
          fi

      - name: Generate
        run: python scripts/refresh_holidays.py --year ${{ steps.year.outputs.year }}

      - name: Open PR
        uses: peter-evans/create-pull-request@v6
        with:
          commit-message: "chore: refresh holidays for ${{ steps.year.outputs.year }}"
          branch: "chore/holidays-${{ steps.year.outputs.year }}"
          delete-branch: true
          title: "chore: refresh holidays for ${{ steps.year.outputs.year }}"
          labels: data, needs-review
          body: |
            Automated holiday refresh for **${{ steps.year.outputs.year }}**.

            ### Reviewer checklist

            Cross-check each file against the official exchange calendar:
            - [ ] NASDAQ — https://www.nyse.com/markets/hours-calendars
            - [ ] HKEX — https://www.hkex.com.hk/Services/Trading-hours-and-Severe-Weather-Arrangements/Trading-Hours/Holiday-Schedule
            - [ ] China A-Share — http://www.sse.com.cn/market/
            - [ ] TSE — https://www.jpx.co.jp/english/corporate/about-jpx/calendar/
            - [ ] KRX — https://global.krx.co.kr/contents/GLB/05/0501/0501020000/GLB0501020000.jsp

            ### Half-day verification

            Half-days are not reliably extracted from `exchange_calendars`. Please confirm any `type: half_day` entries and add any missing ones manually before merging:
            - NASDAQ: Day after Thanksgiving, Christmas Eve (usually)
            - HKEX: Lunar New Year's Eve, Christmas Eve (usually)

            **Do not auto-merge.**
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/refresh-holidays.yml
git commit -m "ci: add yearly GitHub Action to refresh holiday calendars"
```

---

## Task 19: README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write README**

```markdown
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

NASDAQ includes the Blue Ocean ATS overnight session (8pm–4am ET, Sun–Thu).

## Data

Market schedules and holiday calendars live in `data/` as YAML and are embedded into the binary via `go:embed`. A GitHub Action runs yearly (November 15) to open a PR generating next-year holidays from [`exchange_calendars`](https://pypi.org/project/exchange-calendars/). PRs require human review before merge.

## Design

See [docs/superpowers/specs/2026-04-17-tradinghour-design.md](docs/superpowers/specs/2026-04-17-tradinghour-design.md).

## License

MIT — see [LICENSE](LICENSE).
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README with usage and supported markets"
```

---

## Final verification

- [ ] **Run full test suite**

Run: `go test -race ./...`
Expected: all packages `ok`.

- [ ] **Run `go vet` and `go build`**

Run: `go vet ./... && go build ./...`
Expected: clean.

- [ ] **Confirm data file layout**

Run: `find data -type f | sort`
Expected: 5 market YAMLs + 5 holiday YAMLs.

---

## Out-of-scope notes for the implementing engineer

- Don't add `DaySchedule.String()` / `MarshalJSON` unless a caller asks — keep the surface minimal.
- Don't auto-generate holidays *inside* Go — the Python script is the one source of truth for regeneration.
- Don't add a `Register(market MarketType, yaml string)` extension API. Post-MVP only.
- If `go test ./...` reports a flake, investigate; DO NOT retry with sleep loops.

## Spec coverage

| Spec section | Covered by tasks |
|---|---|
| 2.1 Supported markets | 5-6, 12, 13, 14, 15 |
| 2.2 Session model per market | 5, 12, 13, 14, 15 |
| 3 Public API | 2, 9, 10, 11 |
| 3.1 Semantics (overnight, TZ) | 9, 10, 16 |
| 4.1 Market YAML format | 5, 7 |
| 4.2 Holiday YAML format | 6, 7 |
| 5 Timezone semantics | 3, 9, 10, 16 (DST tests) |
| 6 Core algorithms | 8, 9, 10, 11 |
| 7 Package layout | 1, 2, 3, 4, 7, 8 |
| 8 GitHub Action | 17, 18 |
| 9 Testing strategy | 9, 12–16 |
