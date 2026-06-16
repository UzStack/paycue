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

// sendUserWebhook to'lov egasining (user) webhook URLiga ma'lumot yuboradi.
// Maxfiy kalit X-API-Key headerda boradi (client haqiqiylikni tekshiradi).
func sendUserWebhook(db *sql.DB, log *zap.Logger, task domain.WebhookTask) bool {
	url, secret, err := repository.GetWebhook(db, task.UserID)
	if err != nil || url == "" {
		log.Info("webhook sozlanmagan", zap.Int64("user", task.UserID))
		return false
	}
	err = WebhookRequest(url, secret, map[string]any{
		"action":         task.Action,
		"amount":         task.Amount,
		"card_id":        task.CardID,
		"transaction_id": task.TransID,
	}, log, 1)
	return err == nil
}

func CloseTransactionWorker(ctx context.Context, db *sql.DB, log *zap.Logger, cfg *config.Config) {
	for {
		select {
		case <-ctx.Done():
			log.Info("Worker stop transaction")
			return
		default:
			transactions, err := repository.GetOldTransactions(db, cfg.TimeoutMins)
			if err != nil {
				log.Error("old transactions close error", zap.Error(err))
			}
			for _, task := range transactions {
				webhookStatus := sendUserWebhook(db, log, task)
				if err := repository.ConfirmTransaction(db, task.TransID, webhookStatus); err != nil {
					log.Info("transaction cancel error", zap.Error(err))
				} else {
					log.Info("transaction cancel", zap.Int64("amount", task.Amount), zap.String("transaction_id", task.TransID), zap.Bool("webhookStatus", webhookStatus))
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
		case t, ok := <-tasks:
			if !ok {
				continue
			}
			task, ok := t.Paylod().(domain.WebhookTask)
			if !ok {
				continue
			}
			webhookStatus := sendUserWebhook(db, log, task)
			if err := repository.ConfirmTransaction(db, task.TransID, webhookStatus); err != nil {
				log.Error(err.Error())
			}
		}
	}
}
