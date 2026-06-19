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
	AmountInt    int64  // 330000000 (tiyinda: so'm*100 + tiyin)
	Last4        string // "7159" — cartaning oxirgi 4 raqami
	Currency     string // "UZS"
	AmountPretty string
}

// toNumber summani tiyinda qaytaradi (1 so'm = 100 tiyin), kasr (tiyin) qismni
// saqlab: "3 300 000,00" -> 330000000, "1.000,01" -> 100001.
func toNumber(amount string, log *zap.Logger) (int64, bool) {
	// Minglik ajratuvchilar (nuqta yoki bo'shliq) olib tashlanadi,
	// vergul o'nlik nuqtaga aylantiriladi: "3 300 000,00" -> "3300000.00".
	normalized := strings.NewReplacer(".", "", " ", "", ",", ".").Replace(amount)
	whole, frac, _ := strings.Cut(normalized, ".")
	som, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, false
	}
	// Kasr qismni aniq 2 raqamga keltiramiz (tiyin): "" -> 00, "5" -> 50, "123" -> 12.
	switch {
	case len(frac) == 0:
		frac = "00"
	case len(frac) == 1:
		frac += "0"
	case len(frac) > 2:
		frac = frac[:2]
	}
	tiyin, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, false
	}
	return som*100 + tiyin, true
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
