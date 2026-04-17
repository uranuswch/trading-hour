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
