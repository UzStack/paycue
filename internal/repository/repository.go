package repository

import (
	"database/sql"
	"strconv"
	"time"

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
			webhook_attempts INTEGER DEFAULT 0,
			action TEXT DEFAULT '',
			transaction_id TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS card_amount_created_index ON transactions(card_id, amount, created_at);
		CREATE UNIQUE INDEX IF NOT EXISTS transaction_id_index ON transactions(transaction_id);

		CREATE TABLE IF NOT EXISTS webhook_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			transaction_id TEXT DEFAULT '',
			card_id INTEGER DEFAULT 0,
			action TEXT DEFAULT '',
			url TEXT DEFAULT '',
			amount INTEGER DEFAULT 0,
			attempts INTEGER DEFAULT 0,
			success BOOLEAN DEFAULT 0,
			status_code INTEGER DEFAULT 0,
			error TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS webhook_logs_user_index ON webhook_logs(user_id, created_at);

		CREATE TABLE IF NOT EXISTS meta (
			key TEXT PRIMARY KEY,
			value TEXT DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS stats_reports (
			instance_id TEXT PRIMARY KEY,
			version TEXT DEFAULT '',
			os TEXT DEFAULT '',
			arch TEXT DEFAULT '',
			users INTEGER DEFAULT 0,
			telegram_accounts INTEGER DEFAULT 0,
			cards INTEGER DEFAULT 0,
			transactions INTEGER DEFAULT 0,
			transactions_active INTEGER DEFAULT 0,
			transactions_confirmed INTEGER DEFAULT 0,
			transactions_cancelled INTEGER DEFAULT 0,
			webhook_logs INTEGER DEFAULT 0,
			reported_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		panic(err)
	}
	// Eski DB'lar uchun migratsiya: yangi ustunlar bo'lmasa qo'shamiz.
	// (ustun allaqachon bo'lsa SQLite xato qaytaradi — e'tiborsiz qoldiramiz.)
	db.Exec("ALTER TABLE users ADD COLUMN password_hash TEXT DEFAULT ''")
	db.Exec("ALTER TABLE cards ADD COLUMN number TEXT DEFAULT ''")
	db.Exec("ALTER TABLE cards ADD COLUMN owner_name TEXT DEFAULT ''")
	db.Exec("ALTER TABLE transactions ADD COLUMN action TEXT DEFAULT ''")
	db.Exec("ALTER TABLE transactions ADD COLUMN webhook_attempts INTEGER DEFAULT 0")
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

// ConfirmTransaction transactionni yopadi: status=0 qiladi, action (confirm|cancel),
// webhook yetkazilgani holatini va urinishlar sonini saqlaydi.
func ConfirmTransaction(db *sql.DB, transactionID, action string, webhookStatus bool, attempts int) error {
	_, err := db.Exec("UPDATE transactions SET status=0, action=?, webhook_status=?, webhook_attempts=? WHERE transaction_id=?",
		action, webhookStatus, attempts, transactionID)
	return err
}

// WebhookLogsPerUser — har bir foydalanuvchi uchun saqlanadigan webhook loglar maksimumi.
// Yozish (retention) va o'qish (limit) bir xil chegaradan foydalanadi.
const WebhookLogsPerUser = 1000

// CreateWebhookLog bitta webhook yetkazib berish natijasini yozadi va jadval
// cheksiz o'smasligi uchun shu user bo'yicha eng eski ortiqcha loglarni o'chiradi.
func CreateWebhookLog(db *sql.DB, l domain.WebhookLog) error {
	_, err := db.Exec(`INSERT INTO webhook_logs
		(user_id, transaction_id, card_id, action, url, amount, attempts, success, status_code, error)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.UserID, l.TransactionID, l.CardID, l.Action, l.URL, l.Amount, l.Attempts, l.Success, l.StatusCode, l.Error)
	if err != nil {
		return err
	}
	// Retention: eng yangi WebhookLogsPerUser tadan tashqaridagilarni o'chiramiz.
	db.Exec(`DELETE FROM webhook_logs WHERE user_id=? AND id NOT IN (
		SELECT id FROM webhook_logs WHERE user_id=? ORDER BY id DESC LIMIT ?)`,
		l.UserID, l.UserID, WebhookLogsPerUser)
	return nil
}

// ListWebhookLogsByUser foydalanuvchining webhook loglarini (oxirgilari birinchi)
// qaytaradi. Cheksiz o'smasligi uchun limit bilan cheklangan.
func ListWebhookLogsByUser(db *sql.DB, userID int64, limit int) ([]domain.WebhookLog, error) {
	rows, err := db.Query(`SELECT id, user_id, COALESCE(transaction_id,''), card_id, COALESCE(action,''),
		COALESCE(url,''), amount, attempts, success, status_code, COALESCE(error,''), created_at
		FROM webhook_logs WHERE user_id=? ORDER BY id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.WebhookLog
	for rows.Next() {
		var l domain.WebhookLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.TransactionID, &l.CardID, &l.Action,
			&l.URL, &l.Amount, &l.Attempts, &l.Success, &l.StatusCode, &l.Error, &l.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, nil
}

func DeleteTransaction(db *sql.DB, transactionID string) error {
	_, err := db.Exec("DELETE FROM transactions WHERE transaction_id=?", transactionID)
	return err
}

// DeleteTransactionByID transactionni raqamli id bo'yicha o'chiradi.
func DeleteTransactionByID(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM transactions WHERE id=?", id)
	return err
}

// TransactionOwner transaction egasini (user_id) qaytaradi (ownership tekshirish uchun).
func TransactionOwner(db *sql.DB, id int64) (int64, error) {
	var userID int64
	err := db.QueryRow(`SELECT t.user_id FROM transactions tr
		JOIN cards c ON c.id = tr.card_id
		JOIN telegram_accounts t ON t.id = c.telegram_account_id
		WHERE tr.id=?`, id).Scan(&userID)
	return userID, err
}

// ListTransactionsByUser foydalanuvchi transactionlarini carta ma'lumoti va
// hisoblangan holat (state) bilan qaytaradi. timeoutMins active/expired farqi uchun.
func ListTransactionsByUser(db *sql.DB, userID int64, timeoutMins int) ([]domain.Transaction, error) {
	rows, err := db.Query(`
		SELECT tr.id, tr.card_id, tr.amount, tr.status, tr.webhook_status, COALESCE(tr.webhook_attempts,0),
		       COALESCE(tr.action,''), tr.transaction_id, tr.created_at,
		       COALESCE(c.number,''), c.last4, COALESCE(c.owner_name,'')
		FROM transactions tr
		JOIN cards c ON c.id = tr.card_id
		JOIN telegram_accounts t ON t.id = c.telegram_account_id
		WHERE t.user_id = ?
		ORDER BY tr.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cutoff := time.Now().Add(-time.Duration(timeoutMins) * time.Minute)
	var list []domain.Transaction
	for rows.Next() {
		var tr domain.Transaction
		if err := rows.Scan(&tr.ID, &tr.CardID, &tr.Amount, &tr.Status, &tr.WebhookStatus, &tr.WebhookAttempts,
			&tr.Action, &tr.TransactionID, &tr.CreatedAt,
			&tr.CardNumber, &tr.CardLast4, &tr.CardOwner); err != nil {
			return nil, err
		}
		tr.State = transactionState(tr, cutoff)
		list = append(list, tr)
	}
	return list, nil
}

// GetTransactionByTransID transaction_id (UUID) bo'yicha bitta transactionni carta
// ma'lumoti va holati bilan qaytaradi (public to'lov sahifasi uchun, user scopesiz).
func GetTransactionByTransID(db *sql.DB, transID string, timeoutMins int) (*domain.Transaction, error) {
	var tr domain.Transaction
	err := db.QueryRow(`
		SELECT tr.id, tr.card_id, tr.amount, tr.status, tr.webhook_status, COALESCE(tr.webhook_attempts,0),
		       COALESCE(tr.action,''), tr.transaction_id, tr.created_at,
		       COALESCE(c.number,''), c.last4, COALESCE(c.owner_name,'')
		FROM transactions tr
		JOIN cards c ON c.id = tr.card_id
		WHERE tr.transaction_id = ?`, transID).
		Scan(&tr.ID, &tr.CardID, &tr.Amount, &tr.Status, &tr.WebhookStatus, &tr.WebhookAttempts,
			&tr.Action, &tr.TransactionID, &tr.CreatedAt,
			&tr.CardNumber, &tr.CardLast4, &tr.CardOwner)
	if err != nil {
		return nil, err
	}
	tr.State = transactionState(tr, time.Now().Add(-time.Duration(timeoutMins)*time.Minute))
	return &tr, nil
}

// transactionState ko'rsatish uchun holat hisoblaydi: active | confirmed | cancelled | expired.
func transactionState(tr domain.Transaction, cutoff time.Time) string {
	if tr.Status {
		// hali ochiq — timeout ichida bo'lsa active, aks holda close worker hali yetib bormagan (expired).
		if tr.CreatedAt.After(cutoff) {
			return "active"
		}
		return "expired"
	}
	switch tr.Action {
	case "confirm":
		return "confirmed"
	case "cancel":
		return "cancelled"
	default:
		return "cancelled" // eski yozuvlar (action saqlanmagan)
	}
}

// minutesArg datetime('now', '-N minutes') uchun argument tayyorlaydi.
func minutesArg(mins int) string {
	return "-" + strconv.Itoa(mins) + " minutes"
}

// ---- Stats / telemetriya ----

// GetOrCreateInstanceID bu instance uchun barqaror anonim identifikator qaytaradi
// (meta jadvalida saqlanadi, maxfiy emas — faqat hisobotlarni ajratish uchun).
func GetOrCreateInstanceID(db *sql.DB) (string, error) {
	var id string
	if err := db.QueryRow("SELECT value FROM meta WHERE key='instance_id'").Scan(&id); err == nil && id != "" {
		return id, nil
	}
	id = uuid.New().String()
	_, err := db.Exec("INSERT OR REPLACE INTO meta(key, value) VALUES('instance_id', ?)", id)
	return id, err
}

// LocalStats bu instance bo'yicha anonim agregat sanoqlarni yig'adi (PII yo'q).
func LocalStats(db *sql.DB) domain.StatsReport {
	count := func(q string) int {
		var n int
		db.QueryRow(q).Scan(&n)
		return n
	}
	return domain.StatsReport{
		Users:                 count("SELECT count(*) FROM users"),
		TelegramAccounts:      count("SELECT count(*) FROM telegram_accounts"),
		Cards:                 count("SELECT count(*) FROM cards"),
		Transactions:          count("SELECT count(*) FROM transactions"),
		TransactionsActive:    count("SELECT count(*) FROM transactions WHERE status=1"),
		TransactionsConfirmed: count("SELECT count(*) FROM transactions WHERE status=0 AND action='confirm'"),
		TransactionsCancelled: count("SELECT count(*) FROM transactions WHERE status=0 AND action!='confirm'"),
		WebhookLogs:           count("SELECT count(*) FROM webhook_logs"),
	}
}

// SaveStatsReport instance hisobotini saqlaydi (instance_id bo'yicha upsert — oxirgi snapshot).
func SaveStatsReport(db *sql.DB, r domain.StatsReport) error {
	_, err := db.Exec(`INSERT INTO stats_reports
		(instance_id, version, os, arch, users, telegram_accounts, cards, transactions,
		 transactions_active, transactions_confirmed, transactions_cancelled, webhook_logs, reported_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP)
		ON CONFLICT(instance_id) DO UPDATE SET
		 version=excluded.version, os=excluded.os, arch=excluded.arch,
		 users=excluded.users, telegram_accounts=excluded.telegram_accounts, cards=excluded.cards,
		 transactions=excluded.transactions, transactions_active=excluded.transactions_active,
		 transactions_confirmed=excluded.transactions_confirmed, transactions_cancelled=excluded.transactions_cancelled,
		 webhook_logs=excluded.webhook_logs, reported_at=CURRENT_TIMESTAMP`,
		r.InstanceID, r.Version, r.OS, r.Arch, r.Users, r.TelegramAccounts, r.Cards, r.Transactions,
		r.TransactionsActive, r.TransactionsConfirmed, r.TransactionsCancelled, r.WebhookLogs)
	return err
}

// AggregateStats barcha instance hisobotlari bo'yicha jamlanma qaytaradi.
func AggregateStats(db *sql.DB) (domain.StatsAggregate, error) {
	a := domain.StatsAggregate{Versions: map[string]int{}}
	err := db.QueryRow(`SELECT count(*),
		COALESCE(sum(users),0), COALESCE(sum(telegram_accounts),0), COALESCE(sum(cards),0),
		COALESCE(sum(transactions),0), COALESCE(sum(transactions_active),0),
		COALESCE(sum(transactions_confirmed),0), COALESCE(sum(transactions_cancelled),0),
		COALESCE(sum(webhook_logs),0)
		FROM stats_reports`).Scan(&a.Instances, &a.Users, &a.TelegramAccounts, &a.Cards,
		&a.Transactions, &a.TransactionsActive, &a.TransactionsConfirmed, &a.TransactionsCancelled, &a.WebhookLogs)
	if err != nil {
		return a, err
	}
	rows, err := db.Query("SELECT COALESCE(version,''), count(*) FROM stats_reports GROUP BY version ORDER BY count(*) DESC")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var v string
			var n int
			if rows.Scan(&v, &n) == nil {
				if v == "" {
					v = "noma'lum"
				}
				a.Versions[v] = n
			}
		}
	}
	return a, nil
}
