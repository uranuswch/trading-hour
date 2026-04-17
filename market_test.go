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
