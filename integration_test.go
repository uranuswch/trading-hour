package tradinghour

import (
	"testing"
	"time"
)

// Round-trip: for every phase in a day's timeline, IsOpen at Start should
// return open with the phase's session, and IsOpen at End should return closed
// (or open-on-next-phase if phases are contiguous).
func TestTimelineIsOpenConsistency(t *testing.T) {
	markets := []struct {
		m    MarketType
		date time.Time
	}{
		// Use a known non-holiday Monday for each market.
		// 2026-03-02 is a Korean holiday (Samiljeol observed), so KRX uses 2026-03-09.
		{MarketNASDAQ, time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)},
		{MarketHKEX, time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)},
		{MarketChinaAShare, time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)},
		{MarketTSE, time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)},
		{MarketKRX, time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range markets {
		m, d := tc.m, tc.date
		ds, err := Timeline(d, m)
		if err != nil {
			t.Fatalf("%s Timeline: %v", m, err)
		}
		for i, p := range ds.Phases {
			st, _ := IsOpen(p.Start.Unix(), m)
			if !st.Open || st.Session != p.Session {
				t.Errorf("%s phase[%d] start: got (%v, %v), want (true, %v)", m, i, st.Open, st.Session, p.Session)
			}
			// One nanosecond before End: still in this phase.
			st, _ = IsOpen(p.End.Add(-time.Nanosecond).Unix(), m)
			if !st.Open || st.Session != p.Session {
				t.Errorf("%s phase[%d] end-1ns: got (%v, %v), want (true, %v)", m, i, st.Open, st.Session, p.Session)
			}
		}
	}
}

// DST spring-forward: on 2026-03-08, NASDAQ clocks jump from 02:00 -> 03:00 EST -> EDT.
// IsOpen at 04:00 Sun EDT (overnight runs through Mon 04:00) and Mon 09:30 EDT should
// correctly reflect the post-DST offsets.
func TestNASDAQSpringForward(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	mon := time.Date(2026, 3, 9, 9, 30, 0, 0, ny)
	st, _ := IsOpen(mon.Unix(), MarketNASDAQ)
	if !st.Open || st.Session != SessionRegular {
		t.Errorf("Mon 09:30 EDT: got (%v, %v), want (true, regular)", st.Open, st.Session)
	}
	// Sunday overnight ends Mon 04:00 EDT. At 03:59 EDT Mon, still overnight.
	at := time.Date(2026, 3, 9, 3, 59, 0, 0, ny)
	st, _ = IsOpen(at.Unix(), MarketNASDAQ)
	if !st.Open || st.Session != SessionOvernight {
		t.Errorf("Mon 03:59 EDT: got (%v, %v), want (true, overnight)", st.Open, st.Session)
	}
}

// DST fall-back: on 2026-11-01, NASDAQ clocks shift 02:00 EDT -> 01:00 EST.
func TestNASDAQFallBack(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	mon := time.Date(2026, 11, 2, 9, 30, 0, 0, ny)
	st, _ := IsOpen(mon.Unix(), MarketNASDAQ)
	if !st.Open || st.Session != SessionRegular {
		t.Errorf("Mon 09:30 post-fallback: got (%v, %v)", st.Open, st.Session)
	}
}
