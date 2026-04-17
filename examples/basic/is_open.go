package main

import (
	"fmt"
	"log"
	"time"

	tradinghour "github.com/uranuswch/trading-hour"
)

func main() {
	// Current time in Unix seconds
	now := time.Now().Unix()

	// Check if multiple markets are open
	markets := []tradinghour.MarketType{
		tradinghour.MarketNASDAQ,
		tradinghour.MarketHKEX,
		tradinghour.MarketChinaAShare,
		tradinghour.MarketTSE,
		tradinghour.MarketKRX,
		tradinghour.MarketCME,
		tradinghour.MarketFX,
		tradinghour.MarketICE,
		tradinghour.MarketFXCMUKOil,
		tradinghour.MarketFXCMUSOil,
		tradinghour.MarketMetals,
		tradinghour.MarketRates,
	}

	fmt.Println("=== Market Status Check ===")
	fmt.Printf("Time: %s\n\n", time.Now().Format(time.RFC1123))

	for _, market := range markets {
		status, err := tradinghour.IsOpen(now, market)
		if err != nil {
			log.Printf("Error checking %s: %v\n", market, err)
			continue
		}

		state := "CLOSED"
		if status.Open {
			state = "OPEN"
		}
		fmt.Printf("%-15s: %s (session: %s)\n", market, state, status.Session)
	}
}

// Example output:
// === Market Status Check ===
// Time: Mon, 17 Apr 2026 19:15:32 CST
//
// NASDAQ         : CLOSED (session: closed)
// HKEX           : CLOSED (session: closed)
// ChinaAShare    : CLOSED (session: closed)
// TSE            : CLOSED (session: closed)
// KRX            : CLOSED (session: closed)
