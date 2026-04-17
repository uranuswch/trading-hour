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
		SessionPostMarket, SessionOvernight,
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
