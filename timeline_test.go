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
	utc := time.Date(2026, 3, 2, 2, 0, 0, 0, time.UTC)
	ds, _ := Timeline(utc, MarketNASDAQ)
	if ds.Date.Month() != 3 || ds.Date.Day() != 2 {
		t.Errorf("Date = %v, want 2026-03-02", ds.Date)
	}
}
