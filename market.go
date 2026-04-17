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
