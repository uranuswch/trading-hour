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
