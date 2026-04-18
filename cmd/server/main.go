package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	th "github.com/uranuswch/trading-hour"
)

var allMarkets = []th.MarketType{
	th.MarketNASDAQ,
	th.MarketHKEX,
	th.MarketChinaAShare,
	th.MarketTSE,
	th.MarketKRX,
	th.MarketFX,
	th.MarketCME,
	th.MarketICE,
	th.MarketFXCMUKOil,
	th.MarketFXCMUSOil,
	th.MarketRates,
	th.MarketMetals,
}

type statusItem struct {
	Market  string `json:"market"`
	Open    bool   `json:"open"`
	Session string `json:"session"`
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Unix()
	items := make([]statusItem, 0, len(allMarkets))
	for _, m := range allMarkets {
		s, err := th.IsOpen(now, m)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		items = append(items, statusItem{
			Market:  string(m),
			Open:    s.Open,
			Session: string(s.Session),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", handleStatus)
	// web/static is resolved relative to cwd; run from the repo root: go run ./cmd/server/
	mux.Handle("/", http.FileServer(http.Dir("web/static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("listening on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
