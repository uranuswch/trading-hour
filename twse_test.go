package tradinghour

import (
	"testing"
	"time"
)

func TestIsOpenTWSE(t *testing.T) {
	cases := []struct {
		name string
		at   string
		want Session
	}{
		{"before pre-open", "2026-03-02T08:29:59+08:00", SessionClosed},
		{"pre-open orders only", "2026-03-02T08:30:00+08:00", SessionClosed},
		{"before regular", "2026-03-02T08:59:59+08:00", SessionClosed},
		{"regular open", "2026-03-02T09:00:00+08:00", SessionRegular},
		{"no lunch break", "2026-03-02T12:00:00+08:00", SessionRegular},
		{"closing auction", "2026-03-02T13:25:00+08:00", SessionRegular},
		{"before regular close", "2026-03-02T13:29:59+08:00", SessionRegular},
		{"regular close", "2026-03-02T13:30:00+08:00", SessionClosed},
		{"odd-lot orders only", "2026-03-02T13:40:00+08:00", SessionClosed},
		{"before postmarket", "2026-03-02T13:59:59+08:00", SessionClosed},
		{"postmarket open", "2026-03-02T14:00:00+08:00", SessionPostMarket},
		{"before postmarket close", "2026-03-02T14:29:59+08:00", SessionPostMarket},
		{"postmarket close", "2026-03-02T14:30:00+08:00", SessionClosed},
		{"overnight closed", "2026-03-02T23:00:00+08:00", SessionClosed},
		{"UTC open", "2026-03-02T01:00:00Z", SessionRegular},
		{"summer UTC open", "2026-07-01T01:00:00Z", SessionRegular},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			at, err := time.Parse(time.RFC3339, tc.at)
			if err != nil {
				t.Fatal(err)
			}
			got, err := IsOpen(at.Unix(), MarketTWSE)
			if err != nil {
				t.Fatal(err)
			}
			want := Status{Open: tc.want != SessionClosed, Session: tc.want, Market: MarketTWSE}
			if got != want {
				t.Errorf("status = %+v, want %+v", got, want)
			}
		})
	}
}

func TestTimelineTWSE(t *testing.T) {
	loc, err := MarketLocation(MarketTWSE)
	if err != nil {
		t.Fatal(err)
	}
	if loc.String() != "Asia/Taipei" {
		t.Fatalf("location = %s, want Asia/Taipei", loc)
	}
	// Timeline uses the supplied Y/M/D even when the instant is the next day in Taipei.
	ds, err := Timeline(time.Date(2026, 3, 2, 23, 45, 0, 0, time.UTC), MarketTWSE)
	if err != nil {
		t.Fatal(err)
	}
	if ds.Market != MarketTWSE || ds.Date.Location() != loc ||
		ds.Date.Format(time.RFC3339) != "2026-03-02T00:00:00+08:00" {
		t.Errorf("unexpected market-local date: %+v", ds)
	}
	want := []Phase{
		{SessionRegular, time.Date(2026, 3, 2, 9, 0, 0, 0, loc), time.Date(2026, 3, 2, 13, 30, 0, 0, loc)},
		{SessionPostMarket, time.Date(2026, 3, 2, 14, 0, 0, 0, loc), time.Date(2026, 3, 2, 14, 30, 0, 0, loc)},
	}
	if len(ds.Phases) != len(want) {
		t.Fatalf("phases = %+v, want %+v", ds.Phases, want)
	}
	for i, p := range ds.Phases {
		if p != want[i] {
			t.Errorf("phase[%d] = %+v, want %+v", i, p, want[i])
		}
	}
}

func TestTWSE2026Calendar(t *testing.T) {
	// Published weekday closures, independently checked against the TWSE calendar:
	// https://www.twse.com.tw/holidaySchedule/holidaySchedule?response=html&queryYear=2026
	holidayDates := []string{
		"2026-01-01", "2026-02-12", "2026-02-13", "2026-02-16", "2026-02-17",
		"2026-02-18", "2026-02-19", "2026-02-20", "2026-02-27", "2026-04-03",
		"2026-04-06", "2026-05-01", "2026-06-19", "2026-09-25", "2026-09-28",
		"2026-10-09", "2026-10-26", "2026-12-25",
	}
	holidays := make(map[string]bool, len(holidayDates))
	for _, date := range holidayDates {
		holidays[date] = true
	}
	for date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC); date.Year() == 2026; date = date.AddDate(0, 0, 1) {
		t.Run(date.Format("2006-01-02"), func(t *testing.T) {
			ds, err := Timeline(date, MarketTWSE)
			if err != nil {
				t.Fatal(err)
			}
			wantHoliday := holidays[date.Format("2006-01-02")]
			if ds.IsHoliday != wantHoliday || ds.IsHalfDay || (ds.HolidayName != "") != wantHoliday {
				t.Errorf("unexpected holiday metadata: %+v", ds)
			}
			wantOpen := !wantHoliday && date.Weekday() != time.Saturday && date.Weekday() != time.Sunday
			wantPhases := 0
			if wantOpen {
				wantPhases = 2
			}
			if len(ds.Phases) != wantPhases {
				t.Errorf("phases = %d, want %d", len(ds.Phases), wantPhases)
			}
			for _, offset := range []time.Duration{10 * time.Hour, 14*time.Hour + 15*time.Minute} {
				st, err := IsOpen(ds.Date.Add(offset).Unix(), MarketTWSE)
				if err != nil {
					t.Fatal(err)
				}
				if st.Open != wantOpen {
					t.Errorf("at %s: open = %v, want %v", ds.Date.Add(offset), st.Open, wantOpen)
				}
			}
		})
	}
}

func TestTWSENextOpenClose(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		at        string
		wantOpen  string
		wantClose string
	}{
		{"before open", "2026-03-02 08:59", "2026-03-02 09:00", "2026-03-02 13:30"},
		{"regular", "2026-03-02 12:00", "2026-03-02 14:00", "2026-03-02 13:30"},
		{"session gap", "2026-03-02 13:30", "2026-03-02 14:00", "2026-03-02 14:30"},
		{"postmarket", "2026-03-02 14:00", "2026-03-03 09:00", "2026-03-02 14:30"},
		{"weekend", "2026-03-06 14:30", "2026-03-09 09:00", "2026-03-09 13:30"},
		{"Lunar New Year", "2026-02-11 14:30", "2026-02-23 09:00", "2026-02-23 13:30"},
		{"settlement only", "2026-02-12 10:00", "2026-02-23 09:00", "2026-02-23 13:30"},
		{"April holidays", "2026-04-02 14:30", "2026-04-07 09:00", "2026-04-07 13:30"},
		{"Mid-Autumn and Teachers Day", "2026-09-24 14:30", "2026-09-29 09:00", "2026-09-29 13:30"},
		{"Restoration Day observed", "2026-10-23 14:30", "2026-10-27 09:00", "2026-10-27 13:30"},
		{"Constitution Day", "2026-12-24 14:30", "2026-12-28 09:00", "2026-12-28 13:30"},
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
				got, err := query.fn(at.Unix(), MarketTWSE)
				if err != nil {
					t.Fatalf("%s: %v", query.name, err)
				}
				if got.Format(layout) != query.want || got.Location().String() != "Asia/Taipei" {
					t.Errorf("%s = %s, want %s Asia/Taipei", query.name, got, query.want)
				}
			}
		})
	}
}
