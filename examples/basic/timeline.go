package main

import (
	"fmt"
	"log"
	"time"

	tradinghour "github.com/uranuswch/trading-hour"
)

func main() {
	// Get timeline for NASDAQ on a specific date
	date := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC) // January 5, 2026

	fmt.Println("=== NASDAQ Timeline for January 5, 2026 ===")

	schedule, err := tradinghour.Timeline(date, tradinghour.MarketNASDAQ)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Date: %s\n", schedule.Date.Format("2006-01-02 Monday"))
	fmt.Printf("Market: %s\n", schedule.Market)
	fmt.Printf("Is Holiday: %v\n", schedule.IsHoliday)
	fmt.Printf("Is Half Day: %v\n", schedule.IsHalfDay)

	if schedule.IsHoliday {
		fmt.Printf("Holiday: %s\n", schedule.HolidayName)
	}

	fmt.Println("\nTrading Sessions:")
	for _, phase := range schedule.Phases {
		start := phase.Start.Format("15:04 MST")
		end := phase.End.Format("15:04 MST")
		fmt.Printf("  %-12s: %s → %s\n", phase.Session, start, end)
	}
}

// Example output:
// === NASDAQ Timeline for January 5, 2026 ===
// Date: 2026-01-05 Monday
// Market: NASDAQ
// Is Holiday: false
// Is Half Day: false
//
// Trading Sessions:
//  premarket    : 04:00 EST → 09:30 EST
//  regular      : 09:30 EST → 16:00 EST
//  postmarket   : 16:00 EST → 20:00 EST
//  overnight    : 20:00 EST → 04:00 EST
