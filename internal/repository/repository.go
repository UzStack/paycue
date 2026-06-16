package repository

import (
	"database/sql"
	"strconv"

	"github.com/UzStack/paycue/internal/domain"
	"github.com/google/uuid"
)

func InitTables(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			phone TEXT,
			password_hash TEXT DEFAULT '',
			token TEXT NOT NULL UNIQUE,
			webhook_url TEXT DEFAULT '',
			webhook_secret TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS users_token_index ON users(token);

		CREATE TABLE IF NOT EXISTS telegram_accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			phone TEXT NOT NULL,
			tg_user_id INTEGER DEFAULT 0,
			username TEXT DEFAULT '',
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, phone)
		);

		CREATE TABLE IF NOT EXISTS cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_account_id INTEGER NOT NULL,
			number TEXT DEFAULT '',
			last4 TEXT NOT NULL,
			owner_name TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(telegram_account_id, last4)
		);
		CREATE INDEX IF NOT EXISTS cards_last4_index ON cards(last4);

		CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_id INTEGER NOT NULL,
			amount INTEGER NOT NULL,
			status BOOLEAN DEFAULT 1,
			webhook_status BOOLEAN DEFAULT 0,
			transaction_id TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS card_amount_created_index ON transactions(card_id, amount, created_at);
		CREATE UNIQUE INDEX IF NOT EXISTS transaction_id_index ON transactions(transaction_id);
	`)
	if err != nil {
		panic(err)
	}
	// Eski DB'lar uchun migratsiya: yangi ustunlar bo'lmasa qo'shamiz.
	// (ustun allaqachon bo'lsa SQLite xato qaytaradi — e'tiborsiz qoldiramiz.)
	db.Exec("ALTER TABLE users ADD COLUMN password_hash TEXT DEFAULT ''")
	db.Exec("ALTER TABLE cards ADD COLUMN number TEXT DEFAULT ''")
	db.Exec("ALTER TABLE cards ADD COLUMN owner_name TEXT DEFAULT ''")
}

// ---- Users ----

func CreateUser(db *sql.DB, name, email, phone, passwordHash, token string) (*domain.User, error) {
	res, err := db.Exec("INSERT INTO users(name, email, phone, password_hash, token) VALUES(?, ?, ?, ?, ?)",
		name, email, phone, passwordHash, token)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &domain.User{ID: id, Name: name, Email: email, Phone: phone, Token: token}, nil
}

// GetLoginByIdentifier email yoki phone bo'yicha login uchun token va parol hashini qaytaradi.
func GetLoginByIdentifier(db *sql.DB, identifier string) (token, passwordHash string, err error) {
	err = db.QueryRow(`SELECT token, COALESCE(password_hash,'') FROM users
		WHERE email=? OR phone=? LIMIT 1`, identifier, identifier).Scan(&token, &passwordHash)
	return token, passwordHash, err
}

func GetUserByToken(db *sql.DB, token string) (*domain.User, error) {
	var u domain.User
	err := db.QueryRow(`SELECT id, name, COALESCE(email,''), COALESCE(phone,''), webhook_url
		FROM users WHERE token=?`, token).Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.WebhookURL)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func SetWebhook(db *sql.DB, userID int64, url, secret string) error {
	_, err := db.Exec("UPDATE users SET webhook_url=?, webhook_secret=? WHERE id=?", url, secret, userID)
	return err
}

func GetWebhook(db *sql.DB, userID int64) (string, string, error) {
	var url, secret string
	err := db.QueryRow("SELECT COALESCE(webhook_url,''), COALESCE(webhook_secret,'') FROM users WHERE id=?", userID).Scan(&url, &secret)
	return url, secret, err
}

// ---- Telegram accounts ----

func CreateTelegramAccount(db *sql.DB, userID int64, phone string) (int64, error) {
	res, err := db.Exec(`INSERT INTO telegram_accounts(user_id, phone, status) VALUES(?, ?, 'pending')
		ON CONFLICT(user_id, phone) DO UPDATE SET status='pending'`, userID, phone)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil || id == 0 {
		// ON CONFLICT update bo'lganda LastInsertId 0 qaytishi mumkin.
		err = db.QueryRow("SELECT id FROM telegram_accounts WHERE user_id=? AND phone=?", userID, phone).Scan(&id)
	}
	return id, err
}

func ActivateTelegramAccount(db *sql.DB, id, tgUserID int64, username string) error {
	_, err := db.Exec("UPDATE telegram_accounts SET status='active', tg_user_id=?, username=? WHERE id=?", tgUserID, username, id)
	return err
}

func GetTelegramAccount(db *sql.DB, id int64) (*domain.TelegramAccount, error) {
	var a domain.TelegramAccount
	err := db.QueryRow(`SELECT id, user_id, phone, tg_user_id, COALESCE(username,''), status
		FROM telegram_accounts WHERE id=?`, id).Scan(&a.ID, &a.UserID, &a.Phone, &a.TgUserID, &a.Username, &a.Status)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func ListTelegramAccounts(db *sql.DB, userID int64) ([]domain.TelegramAccount, error) {
	rows, err := db.Query(`SELECT id, user_id, phone, tg_user_id, COALESCE(username,''), status, created_at
		FROM telegram_accounts WHERE user_id=? ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.TelegramAccount
	for rows.Next() {
		var a domain.TelegramAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.Phone, &a.TgUserID, &a.Username, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

// ListActiveTelegramAccounts server start bo'lganda watcherlarni tiklash uchun.
func ListActiveTelegramAccounts(db *sql.DB) ([]domain.TelegramAccount, error) {
	rows, err := db.Query(`SELECT id, user_id, phone, tg_user_id, COALESCE(username,''), status
		FROM telegram_accounts WHERE status='active'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.TelegramAccount
	for rows.Next() {
		var a domain.TelegramAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.Phone, &a.TgUserID, &a.Username, &a.Status); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

// ---- Cards ----

// CreateCard to'liq raqam va egasi ismi bilan carta yaratadi. last4 — raqamdan
// ajratilgan oxirgi 4 raqam (Telegram xabariga moslash uchun).
func CreateCard(db *sql.DB, telegramAccountID int64, number, last4, ownerName string) (*domain.Card, error) {
	res, err := db.Exec("INSERT INTO cards(telegram_account_id, number, last4, owner_name) VALUES(?, ?, ?, ?)",
		telegramAccountID, number, last4, ownerName)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &domain.Card{ID: id, TelegramAccountID: telegramAccountID, Number: number, Last4: last4, OwnerName: ownerName}, nil
}

func ListCardsByUser(db *sql.DB, userID int64) ([]domain.Card, error) {
	rows, err := db.Query(`SELECT c.id, c.telegram_account_id, COALESCE(c.number,''), c.last4, COALESCE(c.owner_name,''), c.created_at
		FROM cards c JOIN telegram_accounts t ON t.id = c.telegram_account_id
		WHERE t.user_id=? ORDER BY c.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Card
	for rows.Next() {
		var c domain.Card
		if err := rows.Scan(&c.ID, &c.TelegramAccountID, &c.Number, &c.Last4, &c.OwnerName, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func GetCard(db *sql.DB, id int64) (*domain.Card, error) {
	var c domain.Card
	err := db.QueryRow(`SELECT id, telegram_account_id, COALESCE(number,''), last4, COALESCE(owner_name,'') FROM cards WHERE id=?`, id).
		Scan(&c.ID, &c.TelegramAccountID, &c.Number, &c.Last4, &c.OwnerName)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteCard cartani va uning transactionlarini o'chiradi.
func DeleteCard(db *sql.DB, cardID int64) error {
	db.Exec("DELETE FROM transactions WHERE card_id=?", cardID)
	_, err := db.Exec("DELETE FROM cards WHERE id=?", cardID)
	return err
}

// DeleteTelegramAccount accountni, uning cartalarini va ularning transactionlarini o'chiradi.
func DeleteTelegramAccount(db *sql.DB, accountID int64) error {
	db.Exec("DELETE FROM transactions WHERE card_id IN (SELECT id FROM cards WHERE telegram_account_id=?)", accountID)
	db.Exec("DELETE FROM cards WHERE telegram_account_id=?", accountID)
	_, err := db.Exec("DELETE FROM telegram_accounts WHERE id=?", accountID)
	return err
}

// CardOwner cartaning egasi (user_id) ni tekshirish uchun.
func CardOwner(db *sql.DB, cardID int64) (int64, error) {
	var userID int64
	err := db.QueryRow(`SELECT t.user_id FROM cards c JOIN telegram_accounts t ON t.id=c.telegram_account_id WHERE c.id=?`, cardID).Scan(&userID)
	return userID, err
}

// PickLeastLoadedCard user cartalari ichidan hozir active transactioni eng kam
// bo'lganini tanlaydi (summa farqini minimallashtirish uchun). Tenglikda kichik id.
func PickLeastLoadedCard(db *sql.DB, userID int64, timeoutMins int) (int64, error) {
	var cardID int64
	err := db.QueryRow(`
		SELECT c.id
		FROM cards c
		JOIN telegram_accounts t ON t.id = c.telegram_account_id
		LEFT JOIN transactions tr
			ON tr.card_id = c.id AND tr.status = 1
			AND tr.created_at BETWEEN datetime('now', ?) AND datetime('now')
		WHERE t.user_id = ?
		GROUP BY c.id
		ORDER BY COUNT(tr.id) ASC, c.id ASC
		LIMIT 1`, minutesArg(timeoutMins), userID).Scan(&cardID)
	if err != nil {
		return 0, err
	}
	return cardID, nil
}

// GetCardByLast4 muayyan telegram account bo'yicha oxirgi 4 raqamga mos cartani topadi.
func GetCardByLast4(db *sql.DB, telegramAccountID int64, last4 string) (*domain.Card, error) {
	var c domain.Card
	err := db.QueryRow(`SELECT id, telegram_account_id, COALESCE(number,''), last4, COALESCE(owner_name,'')
		FROM cards WHERE telegram_account_id=? AND last4=?`, telegramAccountID, last4).
		Scan(&c.ID, &c.TelegramAccountID, &c.Number, &c.Last4, &c.OwnerName)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ---- Transactions ----

func CreateTransaction(db *sql.DB, cardID, amount int64) (string, error) {
	transactionID := uuid.New().String()
	_, err := db.Exec("INSERT INTO transactions(card_id, amount, transaction_id) VALUES(?, ?, ?)", cardID, amount, transactionID)
	if err != nil {
		return "", err
	}
	return transactionID, nil
}

// CheckTransaction berilgan carta uchun shu summa bo'sh (active emas) ekanini tekshiradi.
func CheckTransaction(db *sql.DB, cardID, amount int64, timeoutMins int) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT count(id) FROM transactions WHERE
		card_id=? AND amount=? AND status=1 AND
		created_at BETWEEN datetime('now', ?) AND datetime('now')`,
		cardID, amount, minutesArg(timeoutMins)).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// GetTransaction pul tushganda carta + summa bo'yicha active transactionni topadi.
func GetTransaction(db *sql.DB, cardID, amount int64, timeoutMins int) (string, error) {
	var transID string
	err := db.QueryRow(`SELECT transaction_id FROM transactions WHERE
		card_id=? AND amount=? AND status=1 AND
		created_at BETWEEN datetime('now', ?) AND datetime('now')`,
		cardID, amount, minutesArg(timeoutMins)).Scan(&transID)
	if err != nil {
		return "", err
	}
	return transID, nil
}

// GetOldTransactions muddati o'tgan active transactionlarni egasi ma'lumoti bilan qaytaradi.
func GetOldTransactions(db *sql.DB, timeoutMins int) ([]domain.WebhookTask, error) {
	rows, err := db.Query(`SELECT tr.transaction_id, tr.amount, tr.card_id, t.user_id
		FROM transactions tr
		JOIN cards c ON c.id = tr.card_id
		JOIN telegram_accounts t ON t.id = c.telegram_account_id
		WHERE tr.status=1 AND tr.created_at <= datetime('now', ?)`, minutesArg(timeoutMins))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.WebhookTask
	for rows.Next() {
		var task domain.WebhookTask
		if err := rows.Scan(&task.TransID, &task.Amount, &task.CardID, &task.UserID); err != nil {
			return nil, err
		}
		task.Action = "cancel"
		list = append(list, task)
	}
	return list, nil
}

func ConfirmTransaction(db *sql.DB, transactionID string, webhookStatus bool) error {
	_, err := db.Exec("UPDATE transactions SET status=0, webhook_status=? WHERE transaction_id=?", webhookStatus, transactionID)
	return err
}

func DeleteTransaction(db *sql.DB, transactionID string) error {
	_, err := db.Exec("DELETE FROM transactions WHERE transaction_id=?", transactionID)
	return err
}

// minutesArg datetime('now', '-N minutes') uchun argument tayyorlaydi.
func minutesArg(mins int) string {
	return "-" + strconv.Itoa(mins) + " minutes"
}
