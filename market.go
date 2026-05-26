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
//
// Overnight phases (Session == SessionOvernight) are anchored to the next
// trading day, not the calendar day they start on. A Mon 20:00 -> Tue 04:00
// BOAT phase survives when Monday is a closed holiday as long as Tuesday is a
// normal trading day, and a normal Sunday's overnight is dropped when Monday
// is a closed holiday.
func (m *Market) materialize(date time.Time) (phases []Phase, isHoliday bool, isHalfDay bool, name string) {
	y, mo, d := date.Date()
	localMidnight := time.Date(y, mo, d, 0, 0, 0, 0, m.Location)
	weekly := m.WeeklyPhases[int(localMidnight.Weekday())]
	nextDayClosed := m.isClosedHoliday(localMidnight.AddDate(0, 0, 1))

	var dayPhases []compiledPhase
	cd := civilDate{year: y, month: int(mo), day: d}
	if h, ok := m.Holidays[cd]; ok {
		switch h.Type {
		case HolidayClosed:
			isHoliday = true
			name = h.Name
		case HolidayHalfDay:
			isHalfDay = true
			name = h.Name
			dayPhases = withoutOvernight(m.HalfDayPhases)
		}
	} else {
		dayPhases = withoutOvernight(weekly)
	}

	if !nextDayClosed {
		for _, c := range weekly {
			if c.Session == SessionOvernight {
				dayPhases = append(dayPhases, c)
			}
		}
	}

	return instantiateAll(dayPhases, localMidnight, m.Location), isHoliday, isHalfDay, name
}

// isClosedHoliday reports whether the given date is a full-closure holiday for
// this market.
func (m *Market) isClosedHoliday(date time.Time) bool {
	y, mo, d := date.Date()
	cd := civilDate{year: y, month: int(mo), day: d}
	h, ok := m.Holidays[cd]
	return ok && h.Type == HolidayClosed
}

func withoutOvernight(cps []compiledPhase) []compiledPhase {
	out := make([]compiledPhase, 0, len(cps))
	for _, c := range cps {
		if c.Session == SessionOvernight {
			continue
		}
		out = append(out, c)
	}
	return out
}

func instantiateAll(cps []compiledPhase, date time.Time, loc *time.Location) []Phase {
	out := make([]Phase, len(cps))
	for i, c := range cps {
		out[i] = c.instantiate(date, loc)
	}
	return out
}
