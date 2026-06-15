package routes

import (
	"database/sql"
	"net/http"

	"github.com/UzStack/paycue/internal/config"
	"github.com/UzStack/paycue/internal/domain"
	"github.com/UzStack/paycue/internal/http/handlers"
	"go.uber.org/zap"
)

func InitRoutes(mux *http.ServeMux, db *sql.DB, log *zap.Logger, tasks chan domain.Task, cfg *config.Config) {
	handler := handlers.NewHandler(db, log, tasks, cfg)
	mux.HandleFunc("/create/transaction/", handler.HandlerHome)
	mux.HandleFunc("/health/", handler.HealthHandler)
}
