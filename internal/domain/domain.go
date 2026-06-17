package domain

import "time"

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
	Amount        int64     `json:"amount"`
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

// WebhookLog bitta webhook yetkazib berish urinishini (natijasini) yozib boradi.
type WebhookLog struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	TransactionID string    `json:"transaction_id"`
	CardID        int64     `json:"card_id"`
	Action        string    `json:"action"` // confirm | cancel
	URL           string    `json:"url"`
	Amount        int64     `json:"amount"`
	Attempts      int       `json:"attempts"`     // jami urinishlar (1..3)
	Success       bool      `json:"success"`      // oxir-oqibat yetkazildimi
	StatusCode    int       `json:"status_code"`  // oxirgi HTTP javob kodi
	Error         string    `json:"error,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
