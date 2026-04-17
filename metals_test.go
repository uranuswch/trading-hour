package tradinghour

import (
	"testing"
	"time"
)

func TestMetals(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name        string
		timestamp   int64
		wantOpen    bool
		wantSession Session
	}{
		{
			name:        "open during continuous session",
			timestamp:   time.Date(2026, 1, 5, 10, 0, 0, 0, ny).Unix(), // Monday 10am ET
			wantOpen:    true,
			wantSession: SessionContinuous,
		},
		{
			name:        "closed during maintenance gap",
			timestamp:   time.Date(2026, 1, 5, 17, 30, 0, 0, ny).Unix(), // Monday 5:30pm ET (in 17:00-18:00 gap)
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
			name:        "open Sunday evening",
			timestamp:   time.Date(2026, 1, 4, 19, 0, 0, 0, ny).Unix(), // Sunday 7pm ET
			wantOpen:    true,
			wantSession: SessionContinuous,
		},
		{
			name:        "boundary at Sunday open",
			timestamp:   time.Date(2026, 1, 4, 18, 0, 0, 0, ny).Unix(), // Sunday 18:00:00 ET
			wantOpen:    true,
			wantSession: SessionContinuous,
		},
		{
			name:        "boundary at Friday close",
			timestamp:   time.Date(2026, 1, 9, 17, 0, 0, 0, ny).Unix(), // Friday 17:00:00 ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
		{
			name:        "open after maintenance ends",
			timestamp:   time.Date(2026, 1, 5, 18, 0, 0, 0, ny).Unix(), // Monday 18:00:00 ET
			wantOpen:    true,
			wantSession: SessionContinuous,
		},
		{
			name:        "closed on Christmas 2026",
			timestamp:   time.Date(2026, 12, 25, 10, 0, 0, 0, ny).Unix(), // Dec 25 10am ET
			wantOpen:    false,
			wantSession: SessionClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsOpen(tt.timestamp, MarketMetals)
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
