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
	MarketFX          MarketType = "FX"
	MarketCME         MarketType = "CME"
	MarketICE         MarketType = "ICE"
	MarketFXCMUKOil   MarketType = "FXCMUKOil"
	MarketFXCMUSOil   MarketType = "FXCMUSOil"
	MarketRates       MarketType = "Rates"
	MarketMetals      MarketType = "Metals"
)

// Session identifies a market phase at a given instant.
type Session string

const (
	SessionClosed     Session = "closed"
	SessionPreMarket  Session = "premarket"
	SessionRegular    Session = "regular"
	SessionPostMarket Session = "postmarket"
	SessionOvernight  Session = "overnight"
	SessionContinuous Session = "continuous"
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

// ErrNoOpenFound is returned by NextOpen/NextClose when no open phase is found
// within the search horizon.
var ErrNoOpenFound = errors.New("tradinghour: no open phase found within search horizon")

// IsOpen reports the market's status at the given absolute instant.
func IsOpen(unixSec int64, m MarketType) (Status, error) {
	mkt, err := lookup(m)
	if err != nil {
		return Status{}, err
	}
	t := time.Unix(unixSec, 0).In(mkt.Location)
	today, _, _, _ := mkt.materialize(t)
	for _, p := range today {
		if (t.Equal(p.Start) || t.After(p.Start)) && t.Before(p.End) {
			return Status{Open: true, Session: p.Session, Market: m}, nil
		}
	}
	yesterday, _, _, _ := mkt.materialize(t.AddDate(0, 0, -1))
	for _, p := range yesterday {
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

const searchHorizonDays = 15

// NextOpen returns the nearest future instant at which IsOpen transitions to true.
func NextOpen(unixSec int64, m MarketType) (time.Time, error) {
	mkt, err := lookup(m)
	if err != nil {
		return time.Time{}, err
	}
	t := time.Unix(unixSec, 0).In(mkt.Location)
	for i := 0; i < searchHorizonDays; i++ {
		day := time.Date(t.Year(), t.Month(), t.Day()+i, 0, 0, 0, 0, mkt.Location)
		phases, _, _, _ := mkt.materialize(day)
		for _, p := range phases {
			if p.Start.After(t) {
				return p.Start, nil
			}
		}
	}
	return time.Time{}, ErrNoOpenFound
}

// NextClose returns the end of the current open phase, or the end of the next open phase if closed.
func NextClose(unixSec int64, m MarketType) (time.Time, error) {
	mkt, err := lookup(m)
	if err != nil {
		return time.Time{}, err
	}
	t := time.Unix(unixSec, 0).In(mkt.Location)
	for i := -1; i < searchHorizonDays; i++ {
		day := time.Date(t.Year(), t.Month(), t.Day()+i, 0, 0, 0, 0, mkt.Location)
		phases, _, _, _ := mkt.materialize(day)
		for _, p := range phases {
			if p.End.After(t) {
				return p.End, nil
			}
		}
	}
	return time.Time{}, ErrNoOpenFound
}
