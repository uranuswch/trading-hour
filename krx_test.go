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

func TestIsOpenKRXExtendedHours(t *testing.T) {
	cases := []struct {
		name string
		at   string
		want Session
	}{
		{"before old close", "2026-09-11T17:59:59+09:00", SessionPostMarket},
		{"old close", "2026-09-11T18:00:00+09:00", SessionClosed},
		{"before regular close", "2026-09-14T15:29:59+09:00", SessionRegular},
		{"regular close", "2026-09-14T15:30:00+09:00", SessionClosed},
		{"before postmarket", "2026-09-14T15:39:59+09:00", SessionClosed},
		{"closing-price trading", "2026-09-14T15:40:00+09:00", SessionPostMarket},
		{"before continuous trading", "2026-09-14T15:59:59+09:00", SessionPostMarket},
		{"continuous trading", "2026-09-14T16:00:00+09:00", SessionPostMarket},
		{"old close now open", "2026-09-14T18:00:00+09:00", SessionPostMarket},
		{"before extended close", "2026-09-14T19:59:59+09:00", SessionPostMarket},
		{"extended close", "2026-09-14T20:00:00+09:00", SessionClosed},
		{"no overnight spillover", "2026-09-15T00:00:00+09:00", SessionClosed},
		{"following weekday", "2026-09-15T19:00:00+09:00", SessionPostMarket},
		{"UTC extended hours", "2026-09-14T10:00:00Z", SessionPostMarket},
		{"UTC extended close", "2026-09-14T11:00:00Z", SessionClosed},
		{"Saturday closed", "2026-09-19T19:00:00+09:00", SessionClosed},
		{"Sunday closed", "2026-09-20T19:00:00+09:00", SessionClosed},
		{"Chuseok eve closed", "2026-09-24T19:00:00+09:00", SessionClosed},
		{"Chuseok closed", "2026-09-25T19:00:00+09:00", SessionClosed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			at, err := time.Parse(time.RFC3339, tc.at)
			if err != nil {
				t.Fatal(err)
			}
			got, err := IsOpen(at.Unix(), MarketKRX)
			if err != nil {
				t.Fatal(err)
			}
			want := Status{Open: tc.want != SessionClosed, Session: tc.want, Market: MarketKRX}
			if got != want {
				t.Errorf("status = %+v, want %+v", got, want)
			}
		})
	}
}

func TestTimelineKRXExtendedHours(t *testing.T) {
	loc, err := MarketLocation(MarketKRX)
	if err != nil {
		t.Fatal(err)
	}
	if loc.String() != "Asia/Seoul" {
		t.Fatalf("location = %s, want Asia/Seoul", loc)
	}
	cases := []struct {
		date      string
		closeHour int // zero means no trading phases
		holiday   bool
	}{
		{"2026-09-11", 18, false},
		{"2026-09-13", 0, false},
		{"2026-09-14", 20, false},
		{"2026-09-15", 20, false},
		{"2026-09-16", 20, false},
		{"2026-09-17", 20, false},
		{"2026-09-18", 20, false},
		{"2026-09-19", 0, false},
		{"2026-09-20", 0, false},
		{"2026-09-24", 0, true},
		{"2026-09-25", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.date, func(t *testing.T) {
			date, err := time.Parse(time.DateOnly, tc.date)
			if err != nil {
				t.Fatal(err)
			}
			// Keep the requested Y/M/D even when the instant is the next day in Seoul.
			ds, err := Timeline(date.Add(23*time.Hour+45*time.Minute), MarketKRX)
			if err != nil {
				t.Fatal(err)
			}
			if ds.Market != MarketKRX || ds.Date.Location() != loc ||
				ds.Date.Format(time.RFC3339) != tc.date+"T00:00:00+09:00" {
				t.Errorf("unexpected market-local date: %+v", ds)
			}
			if ds.IsHoliday != tc.holiday || ds.IsHalfDay || (ds.HolidayName != "") != tc.holiday {
				t.Errorf("unexpected holiday metadata: %+v", ds)
			}
			var want []Phase
			if tc.closeHour != 0 {
				y, m, d := date.Date()
				want = []Phase{
					{SessionPreMarket, time.Date(y, m, d, 8, 0, 0, 0, loc), time.Date(y, m, d, 9, 0, 0, 0, loc)},
					{SessionRegular, time.Date(y, m, d, 9, 0, 0, 0, loc), time.Date(y, m, d, 15, 30, 0, 0, loc)},
					{SessionPostMarket, time.Date(y, m, d, 15, 40, 0, 0, loc), time.Date(y, m, d, tc.closeHour, 0, 0, 0, loc)},
				}
			}
			if len(ds.Phases) != len(want) {
				t.Fatalf("phases = %+v, want %+v", ds.Phases, want)
			}
			for i, p := range ds.Phases {
				if p != want[i] {
					t.Errorf("phase[%d] = %+v, want %+v", i, p, want[i])
				}
			}
		})
	}
}

func TestKRXNextOpenCloseExtendedHours(t *testing.T) {
	loc, err := MarketLocation(MarketKRX)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		at        string
		wantOpen  string
		wantClose string
	}{
		{"historical session gap", "2026-09-11 15:30", "2026-09-11 15:40", "2026-09-11 18:00"},
		{"historical postmarket", "2026-09-11 17:59", "2026-09-14 08:00", "2026-09-11 18:00"},
		{"cutover weekend", "2026-09-11 18:00", "2026-09-14 08:00", "2026-09-14 09:00"},
		{"session gap", "2026-09-14 15:30", "2026-09-14 15:40", "2026-09-14 20:00"},
		{"closing-price trading", "2026-09-14 15:40", "2026-09-15 08:00", "2026-09-14 20:00"},
		{"continuous trading", "2026-09-14 16:00", "2026-09-15 08:00", "2026-09-14 20:00"},
		{"extended hours", "2026-09-14 18:00", "2026-09-15 08:00", "2026-09-14 20:00"},
		{"extended close", "2026-09-14 20:00", "2026-09-15 08:00", "2026-09-15 09:00"},
		{"weekend", "2026-09-18 20:00", "2026-09-21 08:00", "2026-09-21 09:00"},
		{"Chuseok", "2026-09-23 20:00", "2026-09-28 08:00", "2026-09-28 09:00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const layout = "2006-01-02 15:04"
			at, err := time.ParseInLocation(layout, tc.at, loc)
			if err != nil {
				t.Fatal(err)
			}
			for _, query := range []struct {
				name string
				fn   func(int64, MarketType) (time.Time, error)
				want string
			}{
				{"NextOpen", NextOpen, tc.wantOpen},
				{"NextClose", NextClose, tc.wantClose},
			} {
				got, err := query.fn(at.Unix(), MarketKRX)
				if err != nil {
					t.Fatalf("%s: %v", query.name, err)
				}
				if got.Format(layout) != query.want || got.Location() != loc {
					t.Errorf("%s = %s, want %s Asia/Seoul", query.name, got, query.want)
				}
			}
		})
	}
}
