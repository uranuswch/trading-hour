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
