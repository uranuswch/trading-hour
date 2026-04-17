package tradinghour

import (
	"testing"
	"time"
)

func TestRates(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name        string
		timestamp   int64
		wantOpen    bool
		wantSession Session
	}{
		{
			name:        "open during regular hours",
			timestamp:   time.Date(2026, 1, 5, 10, 0, 0, 0, ny).Unix(), // Monday 10am ET
			wantOpen:    true,
			wantSession: SessionRegular,
		},
		{
			name:        "closed before open",
			timestamp:   time.Date(2026, 1, 5, 7, 0, 0, 0, ny).Unix(), // Monday 7am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "closed after close",
			timestamp:   time.Date(2026, 1, 5, 18, 0, 0, 0, ny).Unix(), // Monday 6pm ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "closed on Saturday",
			timestamp:   time.Date(2026, 1, 3, 10, 0, 0, 0, ny).Unix(), // Saturday 10am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "closed on Sunday",
			timestamp:   time.Date(2026, 1, 4, 10, 0, 0, 0, ny).Unix(), // Sunday 10am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "boundary at open",
			timestamp:   time.Date(2026, 1, 5, 8, 0, 0, 0, ny).Unix(), // Monday 8:00:00 ET
			wantOpen:    true,
			wantSession: SessionRegular,
		},
		{
			name:        "boundary at close",
			timestamp:   time.Date(2026, 1, 5, 17, 0, 0, 0, ny).Unix(), // Monday 17:00:00 ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "closed on New Year's Day 2026",
			timestamp:   time.Date(2026, 1, 1, 10, 0, 0, 0, ny).Unix(), // Jan 1 10am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsOpen(tt.timestamp, MarketRates)
			if err != nil {
				t.Fatal(err)
			}
			if got.Open != tt.wantOpen {
				t.Errorf("IsOpen() Open = %v, want %v", got.Open, tt.wantOpen)
			}
			if got.Session != tt.wantSession {
				t.Errorf("IsOpen() Session = %v, want %v", got.Session, tt.wantSession)
			}
		})
	}
}
