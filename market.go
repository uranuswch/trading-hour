package tradinghour

import "time"

// Market holds all compiled schedule and holiday data for a single market.
type Market struct {
	Type          MarketType
	Location      *time.Location
	WeeklyPhases  [7][]compiledPhase // index = int(time.Weekday); Sunday = 0
	HalfDayPhases []compiledPhase    // nil if market has no half-day support
	Holidays      map[civilDate]holidayEntry
}

// registry is populated by loader.go's init().
var registry = map[MarketType]*Market{}

// lookup returns the registered market or an error.
func lookup(m MarketType) (*Market, error) {
	mkt, ok := registry[m]
	if !ok {
		return nil, ErrUnknownMarket
	}
	return mkt, nil
}

// materialize returns the phases for `date` in this market's local timezone,
// plus holiday flags. The input date's Y/M/D is used; its hour/min are ignored.
func (m *Market) materialize(date time.Time) (phases []Phase, isHoliday bool, isHalfDay bool, name string) {
	y, mo, d := date.Date()
	localMidnight := time.Date(y, mo, d, 0, 0, 0, 0, m.Location)

	cd := civilDate{year: y, month: int(mo), day: d}
	if h, ok := m.Holidays[cd]; ok {
		switch h.Type {
		case HolidayClosed:
			return nil, true, false, h.Name
		case HolidayHalfDay:
			return instantiateAll(m.HalfDayPhases, localMidnight, m.Location), false, true, h.Name
		}
	}

	base := m.WeeklyPhases[int(localMidnight.Weekday())]
	return instantiateAll(base, localMidnight, m.Location), false, false, ""
}

func instantiateAll(cps []compiledPhase, date time.Time, loc *time.Location) []Phase {
	out := make([]Phase, len(cps))
	for i, c := range cps {
		out[i] = c.instantiate(date, loc)
	}
	return out
}
