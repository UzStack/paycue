package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/UzStack/paycue/internal/config"
	"github.com/UzStack/paycue/internal/domain"
	"github.com/UzStack/paycue/internal/repository"
	"go.uber.org/zap"
)

func InitWorker(ctx context.Context, log *zap.Logger, tasks <-chan domain.Task, cfg *config.Config, db *sql.DB) error {
	for range cfg.Workers {
		go Worker(ctx, tasks, log, cfg, db)
	}
	log.Info("Workers running: ", zap.Int("workers", cfg.Workers))
	go CloseTransactionWorker(ctx, db, log, cfg)
	return nil
}

func CloseTransactionWorker(ctx context.Context, db *sql.DB, log *zap.Logger, cfg *config.Config) {
	for {
		select {
		case <-ctx.Done():
			log.Info("Worker stop transaction")
		default:
			transactions, err := repository.GetOldTransactions(db)
			if err != nil {
				log.Error("old transactions close error", zap.Error(err))
			}
			for _, transaction := range transactions {
				webhookStatus := true
				err := WebhookRequest(cfg.WebhookURL, map[string]any{
					"action":         "cancel",
					"amount":         transaction["amount"],
					"transaction_id": transaction["transaction_id"],
				}, log, 1)
				if err != nil {
					webhookStatus = false
				}
				if err := repository.ConfirmTransaction(db, transaction["transaction_id"].(string), webhookStatus); err != nil {
					log.Info("trnsaction cancel error", zap.Error(err))
				} else {
					log.Info("transaction cancel", zap.Any("amount", transaction["amount"]), zap.Any("transaction_id", transaction["transaction_id"]), zap.Bool("webhookStatus", webhookStatus))
				}
			}
			time.Sleep(1 * time.Minute)
		}
	}
}

func Worker(ctx context.Context, tasks <-chan domain.Task, log *zap.Logger, cfg *config.Config, db *sql.DB) error {
	for {
		select {
		case <-ctx.Done():
			log.Info("Worker stop ")
			return nil
		case task, ok := <-tasks:
			if !ok {
				continue
			}
			payload := task.Paylod().(domain.WebhookTask)
			webhookStatus := true
			if err := WebhookRequest(cfg.WebhookURL, map[string]any{
				"action":         "confirm",
				"amount":         payload.Amount,
				"transaction_id": payload.TransID,
			}, log, 1); err != nil {
				webhookStatus = false
				log.Error(err.Error())
			}
			repository.ConfirmTransaction(db, payload.TransID, webhookStatus)
		}
	}
}
