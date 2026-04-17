package tradinghour

import (
	"testing"
	"time"
)

func TestNextOpenWhenClosed(t *testing.T) {
	ny := mustNY(t)
	// Sat 10:00 NY - next open is Sun 20:00 (overnight).
	got, err := NextOpen(time.Date(2026, 3, 7, 10, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 3, 8, 20, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextOpenWhenOpen(t *testing.T) {
	ny := mustNY(t)
	// Mon 11:00 is in regular (09:30-16:00). Next open phase is postmarket 16:00.
	got, err := NextOpen(time.Date(2026, 3, 2, 11, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 3, 2, 16, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextOpenSkipsHoliday(t *testing.T) {
	ny := mustNY(t)
	// Christmas 2026 is Fri. NextOpen on Christmas should skip to Sunday 20:00 (Dec 27).
	got, _ := NextOpen(time.Date(2026, 12, 25, 10, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	want := time.Date(2026, 12, 27, 20, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextCloseWhenOpen(t *testing.T) {
	ny := mustNY(t)
	// Mon 11:00 - inside regular (09:30-16:00). NextClose = 16:00.
	got, err := NextClose(time.Date(2026, 3, 2, 11, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 3, 2, 16, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextCloseWhenClosed(t *testing.T) {
	ny := mustNY(t)
	// Sat 10:00 - market closed. NextClose = end of Sun overnight = Mon 04:00.
	got, _ := NextClose(time.Date(2026, 3, 7, 10, 0, 0, 0, ny).Unix(), MarketNASDAQ)
	want := time.Date(2026, 3, 9, 4, 0, 0, 0, ny)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
