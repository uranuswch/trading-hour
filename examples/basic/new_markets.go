package main

import (
	"fmt"
	"log"
	"time"

	tradinghour "github.com/uranuswch/trading-hour"
)

func main() {
	// Demonstrate the new Rates and Metals markets

	fmt.Println("=== New Markets: Rates and Metals ===\n")

	// Example 1: Check Rates market (interest rate products)
	// Rates uses standard business hours: Mon-Fri 8am-5pm ET
	fmt.Println("1. Rates Market (Interest Rate Products)")
	fmt.Println("   Hours: Mon-Fri 08:00-17:00 ET (business hours)")

	// Check Monday 10am ET
	ny, _ := time.LoadLocation("America/New_York")
	ratesTime := time.Date(2026, 1, 5, 10, 0, 0, 0, ny) // Monday 10am ET
	ratesStatus, err := tradinghour.IsOpen(ratesTime.Unix(), tradinghour.MarketRates)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Monday 10am ET: Open=%v, Session=%s\n\n", ratesStatus.Open, ratesStatus.Session)

	// Example 2: Check Metals market (spot gold/silver)
	// Metals uses 24/5 schedule with daily maintenance: 17:00-18:00 ET Mon-Thu
	fmt.Println("2. Metals Market (Spot Gold/Silver)")
	fmt.Println("   Hours: Sun 18:00 ET → Fri 17:00 ET (24/5)")
	fmt.Println("   Maintenance: 17:00-18:00 ET Mon-Thu")

	// Check Monday 10am ET (should be open)
	metalsTime := time.Date(2026, 1, 5, 10, 0, 0, 0, ny)
	metalsStatus, err := tradinghour.IsOpen(metalsTime.Unix(), tradinghour.MarketMetals)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Monday 10am ET: Open=%v, Session=%s\n", metalsStatus.Open, metalsStatus.Session)

	// Check Monday 5:30pm ET (should be closed - maintenance gap)
	metalsMaintTime := time.Date(2026, 1, 5, 17, 30, 0, 0, ny) // Monday 5:30pm ET
	metalsMaintStatus, err := tradinghour.IsOpen(metalsMaintTime.Unix(), tradinghour.MarketMetals)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Monday 5:30pm ET: Open=%v, Session=%s (maintenance)\n\n", metalsMaintStatus.Open, metalsMaintStatus.Session)

	// Show timeline for Metals
	fmt.Println("3. Metals Timeline (Monday, January 5, 2026)")
	metalsSchedule, err := tradinghour.Timeline(metalsTime, tradinghour.MarketMetals)
	if err != nil {
		log.Fatal(err)
	}

	for _, phase := range metalsSchedule.Phases {
		start := phase.Start.Format("15:04 MST")
		end := phase.End.Format("15:04 MST")
		if end == "00:00 EST" {
			end = "00:00+1 EST" // next day
		}
		fmt.Printf("   %-12s: %s → %s\n", phase.Session, start, end)
	}

	// Example 3: FX market (24/5, no maintenance gaps)
	fmt.Println("\n4. FX Market (Spot Foreign Exchange)")
	fmt.Println("   Hours: Sun 17:00 ET → Fri 17:00 ET (continuous 24/5)")

	// Check Wednesday 3am ET (should be open)
	fxTime := time.Date(2026, 1, 7, 3, 0, 0, 0, ny) // Wednesday 3am ET
	fxStatus, err := tradinghour.IsOpen(fxTime.Unix(), tradinghour.MarketFX)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Wednesday 3am ET: Open=%v, Session=%s\n", fxStatus.Open, fxStatus.Session)
}

// Example output:
// === New Markets: Rates and Metals ===
//
// 1. Rates Market (Interest Rate Products)
//    Hours: Mon-Fri 08:00-17:00 ET (business hours)
//    Monday 10am ET: Open=true, Session=regular
//
// 2. Metals Market (Spot Gold/Silver)
//    Hours: Sun 18:00 ET → Fri 17:00 ET (24/5)
//    Maintenance: 17:00-18:00 ET Mon-Thu
//    Monday 10am ET: Open=true, Session=continuous
//    Monday 5:30pm ET: Open=false, Session=closed (maintenance)
//
// 3. Metals Timeline (Monday, January 5, 2026)
//    continuous   : 00:00 EST → 17:00 EST
//    continuous   : 18:00 EST → 00:00+1 EST
//
// 4. FX Market (Spot Foreign Exchange)
//    Hours: Sun 17:00 ET → Fri 17:00 ET (continuous 24/5)
//    Wednesday 3am ET: Open=true, Session=continuous
