package tradinghour

import (
	"strings"
	"testing"
	"time"
)

func TestRegistryLoaded(t *testing.T) {
	m, err := lookup(MarketNASDAQ)
	if err != nil {
		t.Fatalf("NASDAQ not registered: %v", err)
	}
	if m.Location.String() != "America/New_York" {
		t.Errorf("location = %v", m.Location)
	}
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

func TestParseMarketWeeklyOverrides(t *testing.T) {
	// Overrides are deliberately out of order. Each replaces the full week,
	// while holiday closures and half-days continue to take precedence.
	raw := []byte(`
market: TEST
timezone: Asia/Seoul
weekly_schedule:
  monday: &base
    - {session: regular, start: "09:00", end: "18:00"}
  tuesday: *base
half_day_schedule:
  - {session: regular, start: "09:00", end: "13:00"}
weekly_schedule_overrides:
  - effective_from: "2026-10-05"
    weekly_schedule:
      monday:
        - {session: regular, start: "09:00", end: "21:00"}
  - effective_from: "2026-09-14"
    weekly_schedule:
      monday:
        - {session: regular, start: "09:00", end: "20:00"}
`)
	m, err := parseMarket(raw, "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	m.Holidays[civilDate{2026, 9, 21}] = holidayEntry{Name: "Closure", Type: HolidayClosed}
	m.Holidays[civilDate{2026, 9, 28}] = holidayEntry{Name: "Half day", Type: HolidayHalfDay}
	cases := []struct {
		date      string
		closeHour int
		holiday   bool
		halfDay   bool
		name      string
	}{
		{"2026-09-07", 18, false, false, ""},
		{"2026-09-08", 18, false, false, ""},
		{"2026-09-14", 20, false, false, ""},
		{"2026-09-15", 0, false, false, ""},
		{"2026-09-21", 0, true, false, "Closure"},
		{"2026-09-28", 13, false, true, "Half day"},
		{"2026-10-05", 21, false, false, ""},
		{"2026-10-12", 21, false, false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.date, func(t *testing.T) {
			date, err := time.ParseInLocation(time.DateOnly, tc.date, m.Location)
			if err != nil {
				t.Fatal(err)
			}
			phases, holiday, halfDay, name := m.materialize(date)
			if holiday != tc.holiday || halfDay != tc.halfDay || name != tc.name {
				t.Errorf("holiday metadata = (%v, %v, %q), want (%v, %v, %q)", holiday, halfDay, name, tc.holiday, tc.halfDay, tc.name)
			}
			if tc.closeHour == 0 {
				if len(phases) != 0 {
					t.Errorf("closed date has phases: %+v", phases)
				}
				return
			}
			if len(phases) != 1 {
				t.Fatalf("phases = %+v, want one regular phase", phases)
			}
			want := Phase{SessionRegular, date.Add(9 * time.Hour), date.Add(time.Duration(tc.closeHour) * time.Hour)}
			if phases[0] != want {
				t.Errorf("phase = %+v, want %+v", phases[0], want)
			}
		})
	}
}

func TestParseMarketWeeklyOverrideErrors(t *testing.T) {
	cases := []struct {
		name      string
		overrides string
		wantError string
	}{
		{"missing date", `[{weekly_schedule: {}}]`, "bad effective_from"},
		{"invalid date", `[{effective_from: "2026-02-30", weekly_schedule: {}}]`, "bad effective_from"},
		{"timestamp instead of date", `[{effective_from: "2026-09-14T00:00:00+09:00", weekly_schedule: {}}]`, "bad effective_from"},
		{"missing schedule", `[{effective_from: "2026-09-14"}]`, "weekly_schedule is required"},
		{"duplicate date", `[{effective_from: "2026-09-14", weekly_schedule: {}}, {effective_from: "2026-09-14", weekly_schedule: {}}]`, "duplicate effective_from"},
		{"invalid weekday", `[{effective_from: "2026-09-14", weekly_schedule: {moonday: []}}]`, "unknown weekday"},
		{"invalid start", `[{effective_from: "2026-09-14", weekly_schedule: {monday: [{session: regular, start: "24:00", end: "20:00"}]}}]`, "start:"},
		{"invalid end", `[{effective_from: "2026-09-14", weekly_schedule: {monday: [{session: regular, start: "09:00", end: "25:00"}]}}]`, "end:"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := "market: TEST\ntimezone: Asia/Seoul\nweekly_schedule: {}\nweekly_schedule_overrides: " + tc.overrides
			_, err := parseMarket([]byte(raw), "test.yaml")
			if err == nil || !strings.Contains(err.Error(), tc.wantError) || !strings.Contains(err.Error(), "test.yaml") {
				t.Errorf("error = %v, want %q with source path", err, tc.wantError)
			}
		})
	}
}
