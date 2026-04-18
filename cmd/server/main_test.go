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
}
