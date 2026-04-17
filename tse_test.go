package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenTSE(t *testing.T) {
	jp, _ := time.LoadLocation("Asia/Tokyo")
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
	}{
		{"Mon 08:59 closed", time.Date(2026, 3, 2, 8, 59, 0, 0, jp), false},
		{"Mon 09:00 open", time.Date(2026, 3, 2, 9, 0, 0, 0, jp), true},
		{"Mon 11:30 lunch", time.Date(2026, 3, 2, 11, 30, 0, 0, jp), false},
		{"Mon 12:30 open", time.Date(2026, 3, 2, 12, 30, 0, 0, jp), true},
		{"Mon 15:30 closed", time.Date(2026, 3, 2, 15, 30, 0, 0, jp), false},
		{"Golden Week May 5", time.Date(2026, 5, 5, 10, 0, 0, 0, jp), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, _ := IsOpen(c.local.Unix(), MarketTSE)
			if st.Open != c.wantOpen {
				t.Errorf("got open=%v, want %v", st.Open, c.wantOpen)
			}
		})
	}
}
