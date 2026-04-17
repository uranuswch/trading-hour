// Package tradinghour provides trading-hours queries for global equity markets.
package tradinghour

import (
	"errors"
	"time"
)

// MarketType identifies a supported market.
type MarketType string

const (
	MarketNASDAQ      MarketType = "NASDAQ"
	MarketHKEX        MarketType = "HKEX"
	MarketChinaAShare MarketType = "ChinaAShare"
	MarketTSE         MarketType = "TSE"
	MarketKRX         MarketType = "KRX"
)

// Session identifies a market phase at a given instant.
type Session string

const (
	SessionClosed     Session = "closed"
	SessionPreMarket  Session = "premarket"
	SessionRegular    Session = "regular"
	SessionPostMarket Session = "postmarket"
	SessionOvernight  Session = "overnight"
)

// Status is the result of IsOpen.
type Status struct {
	Open    bool
	Session Session
	Market  MarketType
}

// Phase is one open interval on a given day. End may fall on a later calendar day (overnight).
type Phase struct {
	Session Session
	Start   time.Time
	End     time.Time
}

// DaySchedule is the full timeline for a single market-local date.
type DaySchedule struct {
	Date        time.Time
	Market      MarketType
	Phases      []Phase
	IsHoliday   bool
	IsHalfDay   bool
	HolidayName string
}

// ErrUnknownMarket is returned for any API call with an unknown MarketType.
var ErrUnknownMarket = errors.New("tradinghour: unknown market")

// IsOpen reports the market's status at the given absolute instant.
func IsOpen(unixSec int64, m MarketType) (Status, error) {
	mkt, err := lookup(m)
	if err != nil {
		return Status{}, err
	}
	t := time.Unix(unixSec, 0).In(mkt.Location)
	today, _, _, _ := mkt.materialize(t)
	yesterday, _, _, _ := mkt.materialize(t.AddDate(0, 0, -1))
	for _, p := range append(yesterday, today...) {
		if (t.Equal(p.Start) || t.After(p.Start)) && t.Before(p.End) {
			return Status{Open: true, Session: p.Session, Market: m}, nil
		}
	}
	return Status{Open: false, Session: SessionClosed, Market: m}, nil
}

// Timeline returns the full day schedule for the given date, interpreted in the
// market's local timezone (only Y/M/D of date is used; hour/min/sec/loc are ignored).
func Timeline(date time.Time, m MarketType) (DaySchedule, error) {
	mkt, err := lookup(m)
	if err != nil {
		return DaySchedule{}, err
	}
	y, mo, d := date.Date()
	localMidnight := time.Date(y, mo, d, 0, 0, 0, 0, mkt.Location)
	phases, isHoliday, isHalfDay, name := mkt.materialize(localMidnight)
	return DaySchedule{
		Date:        localMidnight,
		Market:      m,
		Phases:      phases,
		IsHoliday:   isHoliday,
		IsHalfDay:   isHalfDay,
		HolidayName: name,
	}, nil
}

// Placeholder stubs so the API surface exists; real implementations come in later tasks.
func NextOpen(unixSec int64, m MarketType) (time.Time, error)     { panic("not implemented") }
func NextClose(unixSec int64, m MarketType) (time.Time, error)    { panic("not implemented") }
