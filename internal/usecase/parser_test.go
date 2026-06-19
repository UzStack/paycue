package usecase

import (
	"testing"

	"go.uber.org/zap"
)

func TestParseTopUp(t *testing.T) {
	msg := "🎉 To'ldirish\n➕ 1.010,00 UZS\n📍 Click P2P WALLET 2\n💳 HUMOCARD *7159\n🕓 05:36 16.06.2026\n💰 23.378,71 UZS"
	res := ParseTopUp(msg, zap.NewNop())
	if res == nil {
		t.Fatal("natija nil, top_up aniqlanmadi")
	}
	// Summa endi tiyinda: 1.010,00 so'm = 101000 tiyin.
	if res.AmountInt != 101000 {
		t.Errorf("amount: kutilgan 101000, olingan %d", res.AmountInt)
	}
	if res.Last4 != "7159" {
		t.Errorf("last4: kutilgan 7159, olingan %q", res.Last4)
	}
	if res.Currency != "UZS" {
		t.Errorf("currency: kutilgan UZS, olingan %q", res.Currency)
	}
}

func TestParseTopUp_Tiyin(t *testing.T) {
	// Noyob summa tiyin qadamida: 1.000,01 so'm = 100001 tiyin.
	msg := "To'ldirish\n➕ 1.000,01 UZS\n💳 HUMOCARD *7159"
	res := ParseTopUp(msg, zap.NewNop())
	if res == nil {
		t.Fatal("natija nil")
	}
	if res.AmountInt != 100001 {
		t.Errorf("amount: kutilgan 100001, olingan %d", res.AmountInt)
	}
}

func TestParseTopUp_NotTopUp(t *testing.T) {
	if res := ParseTopUp("Oddiy xabar matni", zap.NewNop()); res != nil {
		t.Errorf("top_up bo'lmagan xabar uchun nil kutilgan, olingan %+v", res)
	}
}

func TestParseTopUp_NarrowSpace(t *testing.T) {
	// Telegram ko'pincha narrow no-break space ( ) ishlatadi.
	msg := "To‘ldirish\n➕ 3 300 000,00 UZS\n💳 HUMOCARD *0042"
	res := ParseTopUp(msg, zap.NewNop())
	if res == nil {
		t.Fatal("narrow space bilan natija nil")
	}
	if res.Last4 != "0042" {
		t.Errorf("last4: kutilgan 0042, olingan %q", res.Last4)
	}
}
