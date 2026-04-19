package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	th "github.com/uranuswch/trading-hour"
	webpkg "github.com/uranuswch/trading-hour/web"
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

type phaseItem struct {
	Session string `json:"session"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

type timelineResponse struct {
	Market      string      `json:"market"`
	Date        string      `json:"date"`
	Timezone    string      `json:"timezone"`
	IsHoliday   bool        `json:"isHoliday"`
	IsHalfDay   bool        `json:"isHalfDay"`
	HolidayName string      `json:"holidayName"`
	Phases      []phaseItem `json:"phases"`
}

type nextOpenResponse struct {
	Market string `json:"market"`
	Time   string `json:"time"`
	Local  string `json:"local"`
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

func handleTimeline(w http.ResponseWriter, r *http.Request) {
	market := th.MarketType(r.PathValue("market"))

	var date time.Time
	if ds := r.URL.Query().Get("date"); ds != "" {
		parsed, err := time.Parse("2006-01-02", ds)
		if err != nil {
			http.Error(w, "invalid date: use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		date = parsed
	} else {
		loc, err := th.MarketLocation(market)
		if err == nil {
			date = time.Now().In(loc)
		} else {
			date = time.Now() // unknown market — Timeline will return ErrUnknownMarket below
		}
	}

	sched, err := th.Timeline(date, market)
	if err != nil {
		if errors.Is(err, th.ErrUnknownMarket) {
			http.Error(w, "unknown market", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	phases := make([]phaseItem, len(sched.Phases))
	for i, p := range sched.Phases {
		phases[i] = phaseItem{
			Session: string(p.Session),
			Start:   p.Start.Format(time.RFC3339),
			End:     p.End.Format(time.RFC3339),
		}
	}

	resp := timelineResponse{
		Market:      string(sched.Market),
		Date:        sched.Date.Format("2006-01-02"),
		Timezone:    sched.Date.Location().String(),
		IsHoliday:   sched.IsHoliday,
		IsHalfDay:   sched.IsHalfDay,
		HolidayName: sched.HolidayName,
		Phases:      phases,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleNextOpen(w http.ResponseWriter, r *http.Request) {
	market := th.MarketType(r.PathValue("market"))

	t, err := th.NextOpen(time.Now().Unix(), market)
	if err != nil {
		if errors.Is(err, th.ErrUnknownMarket) {
			http.Error(w, "unknown market", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := nextOpenResponse{
		Market: string(market),
		Time:   t.UTC().Format(time.RFC3339),
		Local:  t.Format("15:04 MST"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", handleStatus)
	mux.HandleFunc("GET /api/timeline/{market}", handleTimeline)
	mux.HandleFunc("GET /api/nextopen/{market}", handleNextOpen)

	// Serve embedded dashboard files — binary is self-contained, no cwd dependency.
	staticFS, err := fs.Sub(webpkg.Static, "static")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to prepare static FS:", err)
		os.Exit(1)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

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
