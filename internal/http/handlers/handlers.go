package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/UzStack/paycue/internal/auth"
	"github.com/UzStack/paycue/internal/config"
	"github.com/UzStack/paycue/internal/domain"
	"github.com/UzStack/paycue/internal/http/middleware"
	"github.com/UzStack/paycue/internal/repository"
	"github.com/UzStack/paycue/internal/telegram"
	"github.com/UzStack/paycue/internal/usecase"
	"go.uber.org/zap"
)

type Handler struct {
	DB  *sql.DB
	Log *zap.Logger
	Cfg *config.Config
	TG  *telegram.Manager
}

func NewHandler(db *sql.DB, log *zap.Logger, cfg *config.Config, tg *telegram.Manager) *Handler {
	return &Handler{DB: db, Log: log, Cfg: cfg, TG: tg}
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, code int, status bool, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(domain.Response{Status: status, Data: data})
}

func ok(w http.ResponseWriter, data any) { writeJSON(w, http.StatusOK, true, data) }

func fail(w http.ResponseWriter, code int, detail string) {
	writeJSON(w, code, false, domain.Detail{Detail: detail})
}

func decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// ---- Health ----

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// ---- Register (public) ----

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, http.StatusBadRequest, "noto'g'ri json")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		fail(w, http.StatusBadRequest, "name majburiy")
		return
	}
	if strings.TrimSpace(in.Email) == "" && strings.TrimSpace(in.Phone) == "" {
		fail(w, http.StatusBadRequest, "email yoki phone dan kamida bittasi majburiy")
		return
	}
	// Parol ixtiyoriy. Berilsa kamida 6 belgi bo'lishi va hashlanishi kerak;
	// berilmasa parolsiz account (login ishlamaydi, faqat token bilan kiriladi).
	hash := ""
	if in.Password != "" {
		if len(in.Password) < 6 {
			fail(w, http.StatusBadRequest, "password kamida 6 ta belgidan iborat bo'lishi kerak")
			return
		}
		var hErr error
		hash, hErr = auth.HashPassword(in.Password)
		if hErr != nil {
			fail(w, http.StatusInternalServerError, "parol hashlanmadi")
			return
		}
	}
	token, err := auth.GenerateToken()
	if err != nil {
		fail(w, http.StatusInternalServerError, "token yaratilmadi")
		return
	}
	user, err := repository.CreateUser(h.DB, in.Name, in.Email, in.Phone, hash, token)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	ok(w, map[string]any{
		"id":    user.ID,
		"name":  user.Name,
		"token": user.Token,
	})
}

// Login email/phone + password orqali tokenni qaytaradi (public).
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Login    string `json:"login"` // email yoki phone
		Password string `json:"password"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, http.StatusBadRequest, "noto'g'ri json")
		return
	}
	in.Login = strings.TrimSpace(in.Login)
	if in.Login == "" || in.Password == "" {
		fail(w, http.StatusBadRequest, "login va password majburiy")
		return
	}
	token, hash, err := repository.GetLoginByIdentifier(h.DB, in.Login)
	if err != nil || hash == "" || !auth.CheckPassword(hash, in.Password) {
		fail(w, http.StatusUnauthorized, "login yoki parol noto'g'ri")
		return
	}
	ok(w, map[string]any{"token": token})
}

// ---- Webhook ----

// GetWebhook joriy webhook URL va secretini qaytaradi.
func (h *Handler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r)
	url, secret, err := repository.GetWebhook(h.DB, user.ID)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	ok(w, map[string]any{"url": url, "secret": secret})
}

func (h *Handler) SetWebhook(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r)
	var in struct {
		URL string `json:"url"`
	}
	if err := decode(r, &in); err != nil || strings.TrimSpace(in.URL) == "" {
		fail(w, http.StatusBadRequest, "url majburiy")
		return
	}
	// Mavjud secretni saqlaymiz, bo'lmasa yangisini yaratamiz.
	_, secret, _ := repository.GetWebhook(h.DB, user.ID)
	if secret == "" {
		secret, _ = auth.GenerateSecret()
	}
	if err := repository.SetWebhook(h.DB, user.ID, in.URL, secret); err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	ok(w, map[string]any{"url": in.URL, "secret": secret})
}

// ---- Telegram ----

func (h *Handler) TelegramSendCode(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r)
	var in struct {
		Phone string `json:"phone"`
	}
	if err := decode(r, &in); err != nil || strings.TrimSpace(in.Phone) == "" {
		fail(w, http.StatusBadRequest, "phone majburiy")
		return
	}
	accountID, err := repository.CreateTelegramAccount(h.DB, user.ID, in.Phone)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	account, err := repository.GetTelegramAccount(h.DB, accountID)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.TG.StartLogin(accountID, *account); err != nil {
		fail(w, http.StatusBadGateway, err.Error())
		return
	}
	ok(w, map[string]any{
		"telegram_account_id": accountID,
		"message":             "Tasdiqlash kodi yuborildi. /api/telegram/verify orqali kodni yuboring.",
	})
}

func (h *Handler) TelegramVerify(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r)
	var in struct {
		TelegramAccountID int64  `json:"telegram_account_id"`
		Code              string `json:"code"`
		Password          string `json:"password"`
	}
	if err := decode(r, &in); err != nil || in.TelegramAccountID == 0 {
		fail(w, http.StatusBadRequest, "telegram_account_id va code majburiy")
		return
	}
	account, err := repository.GetTelegramAccount(h.DB, in.TelegramAccountID)
	if err != nil {
		fail(w, http.StatusNotFound, "account topilmadi")
		return
	}
	if account.UserID != user.ID {
		fail(w, http.StatusForbidden, "bu account sizga tegishli emas")
		return
	}
	need2FA, err := h.TG.SubmitCode(in.TelegramAccountID, in.Code, in.Password)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if need2FA {
		ok(w, map[string]any{
			"need_password": true,
			"message":       "2FA yoqilgan. password bilan qayta /api/telegram/verify yuboring.",
		})
		return
	}
	ok(w, map[string]any{
		"telegram_account_id": in.TelegramAccountID,
		"status":              "active",
		"message":             "Telegram account muvaffaqiyatli ulandi.",
	})
}

func (h *Handler) TelegramList(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r)
	list, err := repository.ListTelegramAccounts(h.DB, user.ID)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	ok(w, list)
}

// ---- Cards ----

func (h *Handler) CardCreate(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r)
	var in struct {
		TelegramAccountID int64  `json:"telegram_account_id"`
		Number            string `json:"number"`     // to'liq carta raqami
		OwnerName         string `json:"owner_name"` // carta egasining ismi
	}
	if err := decode(r, &in); err != nil || in.TelegramAccountID == 0 {
		fail(w, http.StatusBadRequest, "telegram_account_id va number majburiy")
		return
	}
	last4 := lastFourDigits(in.Number)
	if last4 == "" {
		fail(w, http.StatusBadRequest, "number kamida 4 ta raqamdan iborat bo'lishi kerak")
		return
	}
	account, err := repository.GetTelegramAccount(h.DB, in.TelegramAccountID)
	if err != nil || account.UserID != user.ID {
		fail(w, http.StatusForbidden, "telegram account sizga tegishli emas")
		return
	}
	card, err := repository.CreateCard(h.DB, in.TelegramAccountID, in.Number, last4, in.OwnerName)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	ok(w, card)
}

// lastFourDigits matndagi raqamlarning oxirgi 4 tasini qaytaradi (yetmasa "").
func lastFourDigits(s string) string {
	digits := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) < 4 {
		return ""
	}
	return string(digits[len(digits)-4:])
}

func (h *Handler) CardList(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r)
	list, err := repository.ListCardsByUser(h.DB, user.ID)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	ok(w, list)
}

// ---- Transactions ----

func (h *Handler) TransactionCreate(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r)
	var in struct {
		CardID int64 `json:"card_id"`
		Amount int64 `json:"amount"`
	}
	if err := decode(r, &in); err != nil || in.Amount <= 0 {
		fail(w, http.StatusBadRequest, "musbat amount majburiy")
		return
	}

	cardID := in.CardID
	if cardID == 0 {
		// card_id berilmagan — eng kam yuklangan cartani avtomatik tanlaymiz.
		var err error
		cardID, err = repository.PickLeastLoadedCard(h.DB, user.ID, h.Cfg.TimeoutMins)
		if err != nil {
			fail(w, http.StatusBadRequest, "carta topilmadi — avval carta qo'shing")
			return
		}
	} else {
		// Aniq carta berilgan — egasi siz ekanligini tekshiramiz.
		owner, err := repository.CardOwner(h.DB, cardID)
		if err != nil || owner != user.ID {
			fail(w, http.StatusForbidden, "carta sizga tegishli emas")
			return
		}
	}

	amount, transID, err := usecase.CreateTransactionForCard(h.DB, cardID, in.Amount, h.Cfg.TimeoutMins)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Qaysi cartada yaratilgani (raqam/egasi) javobda qaytadi.
	resp := map[string]any{
		"amount":         amount,
		"card_id":        cardID,
		"transaction_id": transID,
	}
	if card, err := repository.GetCard(h.DB, cardID); err == nil {
		resp["card"] = map[string]any{
			"id":         card.ID,
			"number":     card.Number,
			"last4":      card.Last4,
			"owner_name": card.OwnerName,
		}
	}
	ok(w, resp)
}
