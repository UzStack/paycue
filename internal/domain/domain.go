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
	Last4             string    `json:"last4"`
	Label             string    `json:"label,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type Transaction struct {
	ID            int64     `json:"id"`
	CardID        int64     `json:"card_id"`
	Amount        int64     `json:"amount"`
	Status        bool      `json:"status"`
	WebhookStatus bool      `json:"webhook_status"`
	TransactionID string    `json:"transaction_id"`
	CreatedAt     time.Time `json:"created_at"`
}
