package usecase

import (
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// TopUpResult parse natijasi.
type TopUpResult struct {
	Type         string // "top_up"
	AmountRaw    string // "3.300.000,00"
	AmountInt    int64  // 3300000 (butun pul birligi sifatida)
	Last4        string // "7159" — cartaning oxirgi 4 raqami
	Currency     string // "UZS"
	AmountPretty string
}

func toNumber(amount string, log *zap.Logger) (int64, bool) {
	// Minglik ajratuvchilar (nuqta yoki bo'shliq) olib tashlanadi,
	// vergul o'nlik nuqtaga aylantiriladi: "3 300 000,00" / "1.010,00" -> butun son.
	normalized := strings.NewReplacer(".", "", " ", "", ",", ".").Replace(amount)
	if res, err := strconv.ParseFloat(normalized, 64); err == nil {
		return int64(res), true
	}
	return 0, false
}

var (
	// To'ldirish so'zini izlash (case-insensitive). Apostrof turli xil
	// bo'lishi mumkin: ' ’ ‘ ` ʻ ´ yoki umuman bo'lmasligi ham mumkin.
	reTopUpWord = regexp.MustCompile(`(?i)To['’‘` + "`" + `ʻ´]?ldirish`)
	// ➕ 3.300.000,00 UZS  yoki  ➕ 3 300 000,00 UZS
	reAmount = regexp.MustCompile(`➕\s*([\d.\s]+,\d{2})\s*([A-Z]{3})`)
	// 💳 HUMOCARD *7159  (oxirgi 4 raqam)
	reCard = regexp.MustCompile(`\*\s*(\d{4})`)
)

// ParseTopUp matndan To'ldirish operatsiyasi bo'lsa ajratib qaytaradi.
func ParseTopUp(text string, log *zap.Logger) *TopUpResult {
	// Unicode no-break space va boshqalarni oddiy bo'shliqqa
	cleaned := strings.ReplaceAll(text, "\u202f", " ")
	cleaned = strings.ReplaceAll(cleaned, "\u00a0", " ")

	if !reTopUpWord.MatchString(cleaned) {
		return nil
	}
	m := reAmount.FindStringSubmatch(cleaned)
	if len(m) != 3 {
		return nil
	}
	amountStr := m[1]
	currency := m[2]

	val, ok := toNumber(amountStr, log)
	if !ok {
		return nil
	}

	var last4 string
	if c := reCard.FindStringSubmatch(cleaned); len(c) == 2 {
		last4 = c[1]
	}

	return &TopUpResult{
		Type:         "top_up",
		AmountRaw:    amountStr,
		AmountInt:    val,
		Last4:        last4,
		Currency:     currency,
		AmountPretty: amountStr + " " + currency,
	}
}
