package tradinghour

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// timeOfDay is an hour/minute plus an optional +1-day offset, used for phases
// that end on the next calendar day (e.g. NASDAQ BOAT 20:00 -> 04:00+1).
type timeOfDay struct {
	hour, minute, dayOffset int
}

// parseTimeOfDay parses "HH:MM" or "HH:MM+1".
func parseTimeOfDay(s string) (timeOfDay, error) {
	var tod timeOfDay
	if s == "" {
		return tod, fmt.Errorf("tradinghour: empty time string")
	}
	base := s
	if strings.HasSuffix(s, "+1") {
		tod.dayOffset = 1
		base = strings.TrimSuffix(s, "+1")
	}
	if len(base) != 5 || base[2] != ':' {
		return tod, fmt.Errorf("tradinghour: invalid time %q (need HH:MM)", s)
	}
	h, err := strconv.Atoi(base[0:2])
	if err != nil || h < 0 || h > 23 {
		return tod, fmt.Errorf("tradinghour: invalid hour in %q", s)
	}
	m, err := strconv.Atoi(base[3:5])
	if err != nil || m < 0 || m > 59 {
		return tod, fmt.Errorf("tradinghour: invalid minute in %q", s)
	}
	tod.hour = h
	tod.minute = m
	return tod, nil
}

// compiledPhase is a phase with parsed times, ready to be instantiated onto a specific date.
type compiledPhase struct {
	Session Session
	Start   timeOfDay
	End     timeOfDay
}

// instantiate materializes this phase onto the given date in the given location.
// The date argument must already be market-local-midnight.
func (c compiledPhase) instantiate(date time.Time, loc *time.Location) Phase {
	y, m, d := date.Date()
	return Phase{
		Session: c.Session,
		Start:   time.Date(y, m, d+c.Start.dayOffset, c.Start.hour, c.Start.minute, 0, 0, loc),
		End:     time.Date(y, m, d+c.End.dayOffset, c.End.hour, c.End.minute, 0, 0, loc),
	}
}
