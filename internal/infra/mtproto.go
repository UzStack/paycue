package infra

import (
	"context"
	"database/sql"

	"github.com/UzStack/paycue/internal/domain"
	"github.com/UzStack/paycue/internal/repository"
	"github.com/UzStack/paycue/internal/usecase"
	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
)

func Mtproto(ctx context.Context, db *sql.DB, log *zap.Logger, watch_id int64, watch bool, tasks chan domain.Task) error {

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
		Logger:  log.Named("gaps"),
	})

	flow := auth.NewFlow(examples.Terminal{}, auth.SendCodeOptions{})

	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger:        log,
		UpdateHandler: gaps,
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(gaps.Handle),
		},
	})
	if err != nil {
		return err
	}

	d.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return nil
		}

		text := msg.Message
		if text == "" {
			return nil
		}

		var senderID int64

		switch from := msg.FromID.(type) {
		case *tg.PeerUser:
			senderID = from.UserID
		case *tg.PeerChannel, *tg.PeerChat:
			return nil
		default:
			// Shaxsiy (1-1) chatlarda kelgan xabarning FromID maydoni bo'sh
			// bo'ladi — yuboruvchi PeerID orqali aniqlanadi. O'zimiz yuborgan
			// (Out) xabarlarni e'tiborga olmaymiz.
			peer, ok := msg.PeerID.(*tg.PeerUser)
			if !ok || msg.Out {
				return nil
			}
			senderID = peer.UserID
		}

		if !isWatched(senderID, watch_id) {
			return nil
		}

		if res := usecase.ParseTopUp(text, log); res != nil {
			log.Info("To'ldirish aniqlandi",
				zap.String("raw", res.AmountRaw),
				zap.Int64("int", res.AmountInt),
			)
			trans_id, err := repository.GetTransaction(db, res.AmountInt)
			if err != nil {
				log.Info("Transaction topilmadi", zap.Error(err), zap.Int64("amount", res.AmountInt))
				return nil
			}
			log.Info("Transaction topildi", zap.String("id", trans_id), zap.Int64("amount", res.AmountInt))
			tasks <- domain.WebhookTask{
				Amount:  res.AmountInt,
				TransID: trans_id,
			}
		} else {
			log.Debug("Xabar top_up emas", zap.String("text", limit(text, 120)))
		}

		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return errors.Wrap(err, "auth")
		}
		user, err := client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "call self")
		}
		if !watch {
			return nil
		}
		return gaps.Run(ctx, client.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				log.Info("Gaps started")
			},
		})
	})
}

func isWatched(id int64, watch_id int64) bool {
	if watch_id != 0 && id == watch_id {
		return true
	}
	return false
}

func limit(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
