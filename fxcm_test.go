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
		{"Mon 00:59 closed (before open)", time.Date(2026, 3, 2, 0, 59, 0, 0, time.UTC), false, SessionClosed},
		{"Mon 01:00 open",                 time.Date(2026, 3, 2, 1, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		{"Mon 12:00 open",                 time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC), true,  SessionContinuous},
		{"Mon 22:00 closed (maint.)",      time.Date(2026, 3, 2, 22, 0, 0, 0, time.UTC), false, SessionClosed},
		{"Tue 00:59 closed (maint.)",      time.Date(2026, 3, 3, 0, 59, 0, 0, time.UTC), false, SessionClosed},
		{"Tue 01:00 open",                 time.Date(2026, 3, 3, 1, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		{"Fri 21:44 open",                 time.Date(2026, 3, 6, 21, 44, 0, 0, time.UTC), true,  SessionContinuous},
		{"Fri 21:45 closed",               time.Date(2026, 3, 6, 21, 45, 0, 0, time.UTC), false, SessionClosed},
		{"Sat 12:00 closed",               time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC), false, SessionClosed},
		{"Sun 12:00 closed",               time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC), false, SessionClosed},
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
		{"Sun 22:59 closed",             time.Date(2026, 3, 1, 22, 59, 0, 0, time.UTC), false, SessionClosed},
		{"Sun 23:00 open",               time.Date(2026, 3, 1, 23, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		{"Mon 12:00 open",               time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		{"Mon 22:00 closed (maint.)",    time.Date(2026, 3, 2, 22, 0, 0, 0, time.UTC), false, SessionClosed},
		{"Mon 22:59 closed (maint.)",    time.Date(2026, 3, 2, 22, 59, 0, 0, time.UTC), false, SessionClosed},
		{"Mon 23:00 open (post-maint.)", time.Date(2026, 3, 2, 23, 0, 0, 0, time.UTC),  true,  SessionContinuous},
		{"Fri 21:44 open",               time.Date(2026, 3, 6, 21, 44, 0, 0, time.UTC), true,  SessionContinuous},
		{"Fri 21:45 closed",             time.Date(2026, 3, 6, 21, 45, 0, 0, time.UTC), false, SessionClosed},
		{"Sat 12:00 closed",             time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC), false, SessionClosed},
		{"New Year's Day closed",        time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), false, SessionClosed},
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
