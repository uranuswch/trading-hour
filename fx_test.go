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
		{"Mon 12:00 open",        time.Date(2026, 3, 2, 12, 0, 0, 0, ny),  true,  SessionContinuous},
		{"Mon 00:00 open",        time.Date(2026, 3, 2, 0, 0, 0, 0, ny),   true,  SessionContinuous},
		{"Sun 16:59 closed",      time.Date(2026, 3, 1, 16, 59, 0, 0, ny), false, SessionClosed},
		{"Sun 17:00 open",        time.Date(2026, 3, 1, 17, 0, 0, 0, ny),  true,  SessionContinuous},
		{"Sun 23:00 open",        time.Date(2026, 3, 1, 23, 0, 0, 0, ny),  true,  SessionContinuous},
		{"Fri 16:59 open",        time.Date(2026, 3, 6, 16, 59, 0, 0, ny), true,  SessionContinuous},
		{"Fri 17:00 closed",      time.Date(2026, 3, 6, 17, 0, 0, 0, ny),  false, SessionClosed},
		{"Sat 12:00 closed",      time.Date(2026, 3, 7, 12, 0, 0, 0, ny),  false, SessionClosed},
		{"New Year's Day closed", time.Date(2026, 1, 1, 12, 0, 0, 0, ny),  false, SessionClosed},
		{"Christmas Day closed",  time.Date(2026, 12, 25, 12, 0, 0, 0, ny), false, SessionClosed},
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
