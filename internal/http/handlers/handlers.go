package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/UzStack/paycue/internal/config"
	"github.com/UzStack/paycue/internal/domain"
	"github.com/UzStack/paycue/internal/repository"
	"go.uber.org/zap"
)

type Handler struct {
	DB    *sql.DB
	Log   *zap.Logger
	Tasks chan domain.Task
	Cfg   *config.Config
}

func NewHandler(db *sql.DB, log *zap.Logger, tasks chan domain.Task, cfg *config.Config) *Handler {
	return &Handler{
		DB:    db,
		Log:   log,
		Tasks: tasks,
		Cfg:   cfg,
	}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"ok": true,
	})
}

func (h *Handler) HandlerHome(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Amount int64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		fmt.Fprintln(w, err.Error())
	}
	var transaction_id string
	amount := data.Amount
	for {
		if amount-data.Amount > h.Cfg.Limit {
			json.NewEncoder(w).Encode(domain.Response{
				Status: false,
				Data: domain.Detail{
					Detail: "Amount must be less than 100",
				},
			})
			return
		}
		status, err := repository.CheckTransaction(h.DB, amount)
		if err != nil {
			h.Log.Error(err.Error())
			amount += 1
			continue
		}
		if status {
			transaction_id, err = repository.CreateTransaction(h.DB, amount)
			if err != nil {
				fmt.Fprintln(w, err.Error())
			}
			break
		}
		amount += 1
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.Response{
		Status: true,
		Data: struct {
			Amount        int64  `json:"amount"`
			TransactionID string `json:"transaction_id"`
		}{Amount: amount, TransactionID: transaction_id},
	})
}
