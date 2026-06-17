package routes

import (
	"database/sql"
	"net/http"

	"github.com/UzStack/paycue/internal/config"
	"github.com/UzStack/paycue/internal/http/handlers"
	"github.com/UzStack/paycue/internal/http/middleware"
	"github.com/UzStack/paycue/internal/telegram"
	"go.uber.org/zap"
)

func InitRoutes(mux *http.ServeMux, db *sql.DB, log *zap.Logger, cfg *config.Config, tg *telegram.Manager) {
	h := handlers.NewHandler(db, log, cfg, tg)

	// Public
	mux.HandleFunc("GET /health/", h.Health)
	mux.HandleFunc("POST /api/register", h.Register)
	mux.HandleFunc("POST /api/login", h.Login)

	// Token bilan himoyalangan
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return middleware.Auth(db, next)
	}

	mux.HandleFunc("GET /api/webhook", auth(h.GetWebhook))
	mux.HandleFunc("POST /api/webhook", auth(h.SetWebhook))

	mux.HandleFunc("POST /api/telegram/send-code", auth(h.TelegramSendCode))
	mux.HandleFunc("POST /api/telegram/verify", auth(h.TelegramVerify))
	mux.HandleFunc("GET /api/telegram", auth(h.TelegramList))
	mux.HandleFunc("DELETE /api/telegram/{id}", auth(h.TelegramDelete))

	mux.HandleFunc("POST /api/cards", auth(h.CardCreate))
	mux.HandleFunc("GET /api/cards", auth(h.CardList))
	mux.HandleFunc("DELETE /api/cards/{id}", auth(h.CardDelete))

	mux.HandleFunc("POST /api/transactions", auth(h.TransactionCreate))
	mux.HandleFunc("GET /api/transactions", auth(h.TransactionList))
	mux.HandleFunc("DELETE /api/transactions/{id}", auth(h.TransactionDelete))

	// Web UI (statik) — WEB_DIR sozlangan bo'lsa, SPA sifatida xizmat qiladi.
	// /api/* va /health/ aniqroq pattern bo'lgani uchun ular ustun keladi.
	if cfg.WebDir != "" {
		mux.HandleFunc("/", handlers.SPAHandler(cfg.WebDir))
	}
}
