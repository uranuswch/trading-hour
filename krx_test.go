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
		{"Mon 07:59 closed", time.Date(2026, 3, 9, 7, 59, 0, 0, kr), false, SessionClosed},
		{"Mon 08:00 pre", time.Date(2026, 3, 9, 8, 0, 0, 0, kr), true, SessionPreMarket},
		{"Mon 09:00 regular", time.Date(2026, 3, 9, 9, 0, 0, 0, kr), true, SessionRegular},
		{"Mon 15:30 closed", time.Date(2026, 3, 9, 15, 30, 0, 0, kr), false, SessionClosed},
		{"Mon 15:35 gap closed", time.Date(2026, 3, 9, 15, 35, 0, 0, kr), false, SessionClosed},
		{"Mon 15:40 post", time.Date(2026, 3, 9, 15, 40, 0, 0, kr), true, SessionPostMarket},
		{"Mon 18:00 closed", time.Date(2026, 3, 9, 18, 0, 0, 0, kr), false, SessionClosed},
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
