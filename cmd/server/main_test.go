// cmd/server/main_test.go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
