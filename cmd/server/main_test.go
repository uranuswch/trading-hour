// cmd/server/main_test.go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleTimeline_today(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/timeline/NASDAQ", nil)
	req.SetPathValue("market", "NASDAQ")
	w := httptest.NewRecorder()
	handleTimeline(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}
	var resp timelineResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Market != "NASDAQ" {
		t.Errorf("expected NASDAQ, got %s", resp.Market)
	}
	if resp.Timezone != "America/New_York" {
		t.Errorf("unexpected timezone %s", resp.Timezone)
	}
	nyLoc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	wantDate := time.Now().In(nyLoc).Format("2006-01-02")
	if resp.Date != wantDate {
		t.Errorf("expected date %s, got %s", wantDate, resp.Date)
	}
}

func TestHandleTimeline_withDate(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/timeline/HKEX?date=2026-04-19", nil)
	req.SetPathValue("market", "HKEX")
	w := httptest.NewRecorder()
	handleTimeline(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}
	var resp timelineResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Date != "2026-04-19" {
		t.Errorf("expected 2026-04-19, got %s", resp.Date)
	}
}

func TestHandleTimeline_TWSE(t *testing.T) {
	cases := []struct {
		date      string
		isHoliday bool
		phases    []phaseItem
	}{
		{"2026-03-02", false, []phaseItem{
			{Session: "regular", Start: "2026-03-02T09:00:00+08:00", End: "2026-03-02T13:30:00+08:00"},
			{Session: "postmarket", Start: "2026-03-02T14:00:00+08:00", End: "2026-03-02T14:30:00+08:00"},
		}},
		{"2026-02-12", true, []phaseItem{}},
		{"2026-03-07", false, []phaseItem{}},
	}
	for _, tc := range cases {
		t.Run(tc.date, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/timeline/TWSE?date="+tc.date, nil)
			req.SetPathValue("market", "TWSE")
			w := httptest.NewRecorder()
			handleTimeline(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
			}
			var resp timelineResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}
			if resp.Market != "TWSE" || resp.Date != tc.date || resp.Timezone != "Asia/Taipei" {
				t.Errorf("unexpected market, date or timezone: %+v", resp)
			}
			if resp.IsHoliday != tc.isHoliday || resp.IsHalfDay || (resp.HolidayName != "") != tc.isHoliday {
				t.Errorf("unexpected holiday metadata: %+v", resp)
			}
			if resp.Phases == nil || len(resp.Phases) != len(tc.phases) {
				t.Fatalf("phases = %+v, want %+v", resp.Phases, tc.phases)
			}
			for i, phase := range resp.Phases {
				if phase != tc.phases[i] {
					t.Errorf("phase[%d] = %+v, want %+v", i, phase, tc.phases[i])
				}
			}
		})
	}
}

func TestHandleTimeline_unknownMarket(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/timeline/UNKNOWN", nil)
	req.SetPathValue("market", "UNKNOWN")
	w := httptest.NewRecorder()
	handleTimeline(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleTimeline_invalidDate(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/timeline/NASDAQ?date=not-a-date", nil)
	req.SetPathValue("market", "NASDAQ")
	w := httptest.NewRecorder()
	handleTimeline(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleStatus_returnsAllMarkets(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var items []statusItem
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) != len(allMarkets) {
		t.Fatalf("expected %d markets, got %d", len(allMarkets), len(items))
	}
	seen := map[string]bool{}
	for _, item := range items {
		seen[item.Market] = true
	}
	for _, m := range allMarkets {
		if !seen[string(m)] {
			t.Errorf("missing market %s in response", m)
		}
	}
	if !seen["TWSE"] {
		t.Error("missing Taiwan market in response")
	}
}

func TestHandleNextOpen_returnsTime(t *testing.T) {
	for _, market := range []string{"HKEX", "TWSE"} {
		t.Run(market, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/nextopen/"+market, nil)
			req.SetPathValue("market", market)
			w := httptest.NewRecorder()
			handleNextOpen(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
			}
			var resp nextOpenResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}
			if resp.Market != market {
				t.Errorf("expected %s, got %s", market, resp.Market)
			}
			if resp.Time == "" {
				t.Error("empty time field")
			}
			if resp.Local == "" {
				t.Error("empty local field")
			}
		})
	}
}

func TestHandleNextOpen_unknownMarket(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/nextopen/UNKNOWN", nil)
	req.SetPathValue("market", "UNKNOWN")
	w := httptest.NewRecorder()
	handleNextOpen(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
