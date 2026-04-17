package tradinghour

import (
	"errors"
	"testing"
)

func TestMarketTypeConstants(t *testing.T) {
	want := map[MarketType]string{
		MarketNASDAQ:      "NASDAQ",
		MarketHKEX:        "HKEX",
		MarketChinaAShare: "ChinaAShare",
		MarketTSE:         "TSE",
		MarketKRX:         "KRX",
		MarketFX:          "FX",
		MarketCME:         "CME",
		MarketICE:         "ICE",
		MarketFXCMUKOil:   "FXCMUKOil",
		MarketFXCMUSOil:   "FXCMUSOil",
		MarketRates:       "Rates",
		MarketMetals:      "Metals",
	}
	for got, s := range want {
		if string(got) != s {
			t.Errorf("MarketType %q != %q", got, s)
		}
	}
}

func TestSessionConstants(t *testing.T) {
	for _, s := range []Session{
		SessionClosed, SessionPreMarket, SessionRegular,
		SessionPostMarket, SessionOvernight, SessionContinuous,
	} {
		if s == "" {
			t.Errorf("session constant is empty")
		}
	}
}

func TestErrUnknownMarket(t *testing.T) {
	if !errors.Is(ErrUnknownMarket, ErrUnknownMarket) {
		t.Fatal("ErrUnknownMarket sentinel must be comparable with errors.Is")
	}
}

func TestErrNoOpenFound(t *testing.T) {
	if !errors.Is(ErrNoOpenFound, ErrNoOpenFound) {
		t.Fatal("ErrNoOpenFound sentinel must be comparable with errors.Is")
	}
}
