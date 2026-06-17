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

// sendUserWebhook to'lov egasining (user) webhook URLiga ma'lumot yuboradi va
// natijani webhook_logs ga yozadi. (success, urinishlar soni) qaytaradi.
// Maxfiy kalit X-API-Key headerda boradi (client haqiqiylikni tekshiradi).
func sendUserWebhook(db *sql.DB, log *zap.Logger, task domain.WebhookTask) (bool, int) {
	url, secret, err := repository.GetWebhook(db, task.UserID)
	if err != nil || url == "" {
		log.Info("webhook sozlanmagan", zap.Int64("user", task.UserID))
		return false, 0
	}
	res := WebhookRequest(url, secret, map[string]any{
		"action":         task.Action,
		"amount":         task.Amount,
		"card_id":        task.CardID,
		"transaction_id": task.TransID,
	}, log)
	// Har bir yetkazib berish natijasini logga yozamiz (Webhook loglar sahifasi uchun).
	if err := repository.CreateWebhookLog(db, domain.WebhookLog{
		UserID:        task.UserID,
		TransactionID: task.TransID,
		CardID:        task.CardID,
		Action:        task.Action,
		URL:           url,
		Amount:        task.Amount,
		Attempts:      res.Attempts,
		Success:       res.Success,
		StatusCode:    res.StatusCode,
		Error:         res.Err,
	}); err != nil {
		log.Error("webhook log yozilmadi", zap.Error(err))
	}
	return res.Success, res.Attempts
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
				webhookStatus, attempts := sendUserWebhook(db, log, task)
				if err := repository.ConfirmTransaction(db, task.TransID, task.Action, webhookStatus, attempts); err != nil {
					log.Info("transaction cancel error", zap.Error(err))
				} else {
					log.Info("transaction cancel", zap.Int64("amount", task.Amount), zap.String("transaction_id", task.TransID), zap.Bool("webhookStatus", webhookStatus), zap.Int("attempts", attempts))
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
			webhookStatus, attempts := sendUserWebhook(db, log, task)
			if err := repository.ConfirmTransaction(db, task.TransID, task.Action, webhookStatus, attempts); err != nil {
				log.Error(err.Error())
			}
		}
	}
}
