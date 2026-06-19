package domain

import (
	"fmt"
	"strconv"
	"time"
)

// Tiyin — pul miqdori ichkarida tiyinda saqlanadi (1 so'm = 100 tiyin).
// Bu noyob summa farqini so'm o'rniga tiyinda berish imkonini beradi
// (masalan 1000.00, 1000.01 ... 1000.99, 1001.00). JSON'ga so'm (o'nlik son)
// bo'lib chiqadi va DB'dan butun son (tiyin) sifatida o'qiladi.
type Tiyin int64

// MarshalJSON tiyinni so'mga (o'nlik son) aylantirib chiqaradi: 100001 -> 1000.01.
func (t Tiyin) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(t)/100, 'f', -1, 64)), nil
}

// Scan DB'dagi butun son (tiyin) qiymatini o'qiydi.
func (t *Tiyin) Scan(v any) error {
	switch n := v.(type) {
	case int64:
		*t = Tiyin(n)
	case nil:
		*t = 0
	default:
		return fmt.Errorf("Tiyin: qo'llab-quvvatlanmaydigan tip %T", v)
	}
	return nil
}

type Response struct {
	Status bool `json:"status"`
	Data   any  `json:"data"`
}

type Detail struct {
	Detail string `json:"detail"`
}

type User struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Token      string    `json:"token,omitempty"`
	WebhookURL string    `json:"webhook_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type TelegramAccount struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Phone     string    `json:"phone"`
	TgUserID  int64     `json:"tg_user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Status    string    `json:"status"` // pending | active
	CreatedAt time.Time `json:"created_at"`
}

type Card struct {
	ID                int64     `json:"id"`
	TelegramAccountID int64     `json:"telegram_account_id"`
	Number            string    `json:"number,omitempty"`     // to'liq carta raqami
	Last4             string    `json:"last4"`                // raqamdan ajratiladi (Telegram xabariga moslash uchun)
	OwnerName         string    `json:"owner_name,omitempty"` // carta egasining ismi
	CreatedAt         time.Time `json:"created_at"`
}

type Transaction struct {
	ID            int64     `json:"id"`
	CardID        int64     `json:"card_id"`
	Amount        Tiyin     `json:"amount"` // tiyinda saqlanadi, JSON'da so'm (o'nlik)
	Status          bool      `json:"status"`           // true=active(ochiq), false=yopilgan
	WebhookStatus   bool      `json:"webhook_status"`   // webhook yetkazildimi
	WebhookAttempts int       `json:"webhook_attempts"` // nechi marta urinilgan
	Action          string    `json:"action,omitempty"` // yopilganda: confirm | cancel
	TransactionID   string    `json:"transaction_id"`
	CreatedAt       time.Time `json:"created_at"`

	// Ro'yxatda (ListTransactionsByUser) to'ldiriladi:
	State      string `json:"state,omitempty"`       // active | confirmed | cancelled | expired
	CardNumber string `json:"card_number,omitempty"` // to'liq carta raqami
	CardLast4  string `json:"card_last4,omitempty"`
	CardOwner  string `json:"card_owner,omitempty"`
}

// StatsReport — bitta instance'ning anonim foydalanish hisoboti (maxfiy ma'lumotsiz).
type StatsReport struct {
	InstanceID            string    `json:"instance_id"`
	Version               string    `json:"version"`
	OS                    string    `json:"os"`
	Arch                  string    `json:"arch"`
	Users                 int       `json:"users"`
	TelegramAccounts      int       `json:"telegram_accounts"`
	Cards                 int       `json:"cards"`
	Transactions          int       `json:"transactions"`
	TransactionsActive    int       `json:"transactions_active"`
	TransactionsConfirmed int       `json:"transactions_confirmed"`
	TransactionsCancelled int       `json:"transactions_cancelled"`
	WebhookLogs           int       `json:"webhook_logs"`
	ReportedAt            time.Time `json:"reported_at,omitempty"`
}

// StatsAggregate — barcha instance'lar bo'yicha jamlanma (kollektor uchun).
type StatsAggregate struct {
	Enabled               bool           `json:"enabled"`
	Instances             int            `json:"instances"`
	Users                 int            `json:"users"`
	TelegramAccounts      int            `json:"telegram_accounts"`
	Cards                 int            `json:"cards"`
	Transactions          int            `json:"transactions"`
	TransactionsActive    int            `json:"transactions_active"`
	TransactionsConfirmed int            `json:"transactions_confirmed"`
	TransactionsCancelled int            `json:"transactions_cancelled"`
	WebhookLogs           int            `json:"webhook_logs"`
	Versions              map[string]int `json:"versions"` // versiya -> instance soni
}

// WebhookLog bitta webhook yetkazib berish urinishini (natijasini) yozib boradi.
type WebhookLog struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	TransactionID string    `json:"transaction_id"`
	CardID        int64     `json:"card_id"`
	Action        string    `json:"action"` // confirm | cancel
	URL           string    `json:"url"`
	Amount        Tiyin     `json:"amount"` // tiyinda saqlanadi, JSON'da so'm (o'nlik)
	Attempts      int       `json:"attempts"`     // jami urinishlar (1..3)
	Success       bool      `json:"success"`      // oxir-oqibat yetkazildimi
	StatusCode    int       `json:"status_code"`  // oxirgi HTTP javob kodi
	Error         string    `json:"error,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
