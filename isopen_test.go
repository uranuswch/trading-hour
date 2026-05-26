package tradinghour

import (
	"testing"
	"time"
)

func mustNY(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestIsOpenNASDAQ(t *testing.T) {
	ny := mustNY(t)
	cases := []struct {
		name      string
		local     time.Time
		wantOpen  bool
		wantSess  Session
	}{
		{"Mon 04:00 premarket start", time.Date(2026, 3, 2, 4, 0, 0, 0, ny), true, SessionPreMarket},
		{"Mon 09:29 premarket end", time.Date(2026, 3, 2, 9, 29, 59, 0, ny), true, SessionPreMarket},
		{"Mon 09:30 regular start", time.Date(2026, 3, 2, 9, 30, 0, 0, ny), true, SessionRegular},
		{"Mon 12:00 regular", time.Date(2026, 3, 2, 12, 0, 0, 0, ny), true, SessionRegular},
		{"Mon 16:00 postmarket start", time.Date(2026, 3, 2, 16, 0, 0, 0, ny), true, SessionPostMarket},
		{"Mon 20:00 overnight start", time.Date(2026, 3, 2, 20, 0, 0, 0, ny), true, SessionOvernight},
		{"Tue 02:00 overnight (spillover)", time.Date(2026, 3, 3, 2, 0, 0, 0, ny), true, SessionOvernight},
		{"Tue 04:00 premarket start", time.Date(2026, 3, 3, 4, 0, 0, 0, ny), true, SessionPreMarket},
		{"Sat any time closed", time.Date(2026, 3, 7, 10, 0, 0, 0, ny), false, SessionClosed},
		{"Fri 21:00 NO overnight (BOAT Sun-Thu)", time.Date(2026, 3, 6, 21, 0, 0, 0, ny), false, SessionClosed},
		{"Sun 21:00 overnight", time.Date(2026, 3, 8, 21, 0, 0, 0, ny), true, SessionOvernight},
		{"Christmas Day closed", time.Date(2026, 12, 25, 11, 0, 0, 0, ny), false, SessionClosed},
		{"Black Friday 12:59 regular (half-day)", time.Date(2026, 11, 27, 12, 59, 0, 0, ny), true, SessionRegular},
		{"Black Friday 13:00 postmarket (half-day)", time.Date(2026, 11, 27, 13, 0, 0, 0, ny), true, SessionPostMarket},
		{"Black Friday 17:00 closed (half-day)", time.Date(2026, 11, 27, 17, 0, 0, 0, ny), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.local.Unix(), MarketNASDAQ)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
			if st.Market != MarketNASDAQ {
				t.Errorf("Market = %v", st.Market)
			}
		})
	}
}

func TestIsOpenUnknownMarket(t *testing.T) {
	_, err := IsOpen(0, MarketType("BOGUS"))
	if err != ErrUnknownMarket {
		t.Errorf("err = %v", err)
	}
}

// NASDAQ BOAT overnight is anchored to the next trading day: it should run on
// a closed-holiday evening when the next day is normal, and it should NOT run
// on a normal evening when the next day is a closed holiday.
// Memorial Day 2026 (Mon 5/25) is the canonical case — closed-holiday on a
// weekday that has an overnight phase in the weekly schedule.
func TestIsOpenNASDAQHolidayOvernight(t *testing.T) {
	ny := mustNY(t)
	cases := []struct {
		name     string
		local    time.Time
		wantOpen bool
		wantSess Session
	}{
		{"Mon Memorial Day 10:00 still closed (day session)", time.Date(2026, 5, 25, 10, 0, 0, 0, ny), false, SessionClosed},
		{"Mon Memorial Day 19:59 just before overnight", time.Date(2026, 5, 25, 19, 59, 0, 0, ny), false, SessionClosed},
		{"Mon Memorial Day 20:00 overnight start", time.Date(2026, 5, 25, 20, 0, 0, 0, ny), true, SessionOvernight},
		{"Mon Memorial Day 23:30 overnight", time.Date(2026, 5, 25, 23, 30, 0, 0, ny), true, SessionOvernight},
		{"Tue post-Memorial-Day 02:00 overnight spillover", time.Date(2026, 5, 26, 2, 0, 0, 0, ny), true, SessionOvernight},
		{"Tue post-Memorial-Day 04:00 premarket start", time.Date(2026, 5, 26, 4, 0, 0, 0, ny), true, SessionPreMarket},
		{"Sun before Memorial Day 21:00 closed (Mon is holiday)", time.Date(2026, 5, 24, 21, 0, 0, 0, ny), false, SessionClosed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := IsOpen(c.local.Unix(), MarketNASDAQ)
			if err != nil {
				t.Fatal(err)
			}
			if st.Open != c.wantOpen || st.Session != c.wantSess {
				t.Errorf("IsOpen = (%v, %v), want (%v, %v)", st.Open, st.Session, c.wantOpen, c.wantSess)
			}
		})
	}
}
