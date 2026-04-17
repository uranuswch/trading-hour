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
