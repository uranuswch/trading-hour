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

// Placeholder stubs so the API surface exists; real implementations come in later tasks.
func IsOpen(unixSec int64, m MarketType) (Status, error)          { panic("not implemented") }
func Timeline(date time.Time, m MarketType) (DaySchedule, error)  { panic("not implemented") }
func NextOpen(unixSec int64, m MarketType) (time.Time, error)     { panic("not implemented") }
func NextClose(unixSec int64, m MarketType) (time.Time, error)    { panic("not implemented") }
