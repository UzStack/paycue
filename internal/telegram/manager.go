package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/UzStack/paycue/internal/config"
	"github.com/UzStack/paycue/internal/domain"
	"github.com/UzStack/paycue/internal/repository"
	"github.com/UzStack/paycue/internal/usecase"

	"github.com/go-faster/errors"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

const humoBotUsername = "HUMOcardbot"
const loginTimeout = 5 * time.Minute

// Manager bir nechta Telegram accountning login va kuzatuvini boshqaradi.
type Manager struct {
	cfg   *config.Config
	db    *sql.DB
	log   *zap.Logger
	tasks chan domain.Task

	rootCtx context.Context

	mu       sync.Mutex
	logins   map[int64]*pendingLogin      // accountID -> davom etayotgan login
	watchers map[int64]context.CancelFunc // accountID -> kuzatuvni to'xtatish
}

type verifyReq struct {
	code     string
	password string
}

type verifyRes struct {
	need2FA bool
	done    bool
	err     error
}

type pendingLogin struct {
	sent     chan error     // SendCode bosqichi natijasi
	verifyCh chan verifyReq // verify endpointidan keluvchi kod/parol
	resCh    chan verifyRes // verify natijasi
}

func NewManager(cfg *config.Config, db *sql.DB, log *zap.Logger, tasks chan domain.Task) *Manager {
	return &Manager{
		cfg:      cfg,
		db:       db,
		log:      log,
		tasks:    tasks,
		logins:   make(map[int64]*pendingLogin),
		watchers: make(map[int64]context.CancelFunc),
	}
}

// RestoreWatchers server ishga tushganda active accountlar uchun kuzatuvni tiklaydi.
func (m *Manager) RestoreWatchers(ctx context.Context) {
	m.rootCtx = ctx
	if err := os.MkdirAll(m.cfg.SessionDir, 0o700); err != nil {
		m.log.Error("session dir yaratilmadi", zap.Error(err))
	}
	accounts, err := repository.ListActiveTelegramAccounts(m.db)
	if err != nil {
		m.log.Error("active accountlarni olishda xato", zap.Error(err))
		return
	}
	for _, a := range accounts {
		m.startWatcher(a, nil)
	}
	m.log.Info("Telegram watcherlar tiklandi", zap.Int("count", len(accounts)))
}

// StartLogin yangi account uchun login boshlaydi va SendCode tugashini kutadi.
func (m *Manager) StartLogin(accountID int64, account domain.TelegramAccount) error {
	m.mu.Lock()
	if _, ok := m.logins[accountID]; ok {
		m.mu.Unlock()
		return errors.New("bu account uchun login allaqachon boshlangan")
	}
	pl := &pendingLogin{
		sent:     make(chan error, 1),
		verifyCh: make(chan verifyReq),
		resCh:    make(chan verifyRes, 1),
	}
	m.logins[accountID] = pl
	m.mu.Unlock()

	m.startWatcher(account, pl)

	select {
	case err := <-pl.sent:
		if err != nil {
			m.removeLogin(accountID)
		}
		return err
	case <-time.After(30 * time.Second):
		m.removeLogin(accountID)
		return errors.New("SendCode vaqti tugadi")
	}
}

// SubmitCode verify endpointidan kelgan kod/parolni login goroutineiga uzatadi.
// need2FA=true bo'lsa parol kerak (qayta SubmitCode chaqiriladi).
func (m *Manager) SubmitCode(accountID int64, code, password string) (bool, error) {
	m.mu.Lock()
	pl, ok := m.logins[accountID]
	m.mu.Unlock()
	if !ok {
		return false, errors.New("bu account uchun aktiv login topilmadi (avval send-code chaqiring)")
	}

	select {
	case pl.verifyCh <- verifyReq{code: code, password: password}:
	case <-time.After(loginTimeout):
		return false, errors.New("login vaqti tugadi")
	}

	select {
	case res := <-pl.resCh:
		if res.need2FA {
			return true, nil
		}
		if res.err != nil {
			return false, res.err
		}
		return false, nil
	case <-time.After(loginTimeout):
		return false, errors.New("login vaqti tugadi")
	}
}

func (m *Manager) removeLogin(accountID int64) {
	m.mu.Lock()
	delete(m.logins, accountID)
	m.mu.Unlock()
}

// startWatcher account uchun client ishga tushiradi (login != nil bo'lsa avval login qiladi).
func (m *Manager) startWatcher(account domain.TelegramAccount, login *pendingLogin) {
	ctx, cancel := context.WithCancel(m.rootCtx)
	m.mu.Lock()
	if old, ok := m.watchers[account.ID]; ok {
		old() // eskisini to'xtatamiz
	}
	m.watchers[account.ID] = cancel
	m.mu.Unlock()

	go func() {
		defer cancel()
		if err := m.run(ctx, account, login); err != nil && ctx.Err() == nil {
			m.log.Error("watcher to'xtadi", zap.Int64("account", account.ID), zap.Error(err))
		}
		m.mu.Lock()
		delete(m.watchers, account.ID)
		m.mu.Unlock()
		m.removeLogin(account.ID)
	}()
}

func (m *Manager) run(ctx context.Context, account domain.TelegramAccount, login *pendingLogin) error {
	botID := new(atomic.Int64)

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
		Logger:  m.log.Named("gaps"),
	})

	sessionPath := filepath.Join(m.cfg.SessionDir, fmt.Sprintf("%d.json", account.ID))
	client := telegram.NewClient(m.cfg.AppID, m.cfg.AppHash, telegram.Options{
		Logger:         m.log.Named("tg").With(zap.Int64("account", account.ID)),
		SessionStorage: &session.FileStorage{Path: sessionPath},
		UpdateHandler:  gaps,
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(gaps.Handle),
		},
	})

	d.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		return m.onMessage(ctx, account, botID.Load(), update)
	})

	return client.Run(ctx, func(ctx context.Context) error {
		if login != nil {
			if err := m.doLogin(ctx, client, account, login); err != nil {
				return err
			}
		} else {
			st, err := client.Auth().Status(ctx)
			if err != nil {
				return errors.Wrap(err, "auth status")
			}
			if !st.Authorized {
				return errors.New("account avtorizatsiya qilinmagan")
			}
		}

		self, err := client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "self")
		}

		// HUMOcardbotni topish, kerak bo'lsa /start bosish.
		m.startHumoBot(ctx, client, botID)

		return gaps.Run(ctx, client.API(), self.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				m.log.Info("kuzatuv boshlandi", zap.Int64("account", account.ID), zap.String("username", self.Username))
			},
		})
	})
}

// doLogin SendCode -> SignIn -> (kerak bo'lsa) 2FA jarayonini boshqaradi.
func (m *Manager) doLogin(ctx context.Context, client *telegram.Client, account domain.TelegramAccount, pl *pendingLogin) error {
	sent, err := client.Auth().SendCode(ctx, account.Phone, auth.SendCodeOptions{})
	if err != nil {
		pl.sent <- err
		return err
	}
	sc, ok := sent.(*tg.AuthSentCode)
	if !ok {
		err := errors.New("kutilmagan SendCode javobi (login.token-based emas)")
		pl.sent <- err
		return err
	}
	pl.sent <- nil // SendCode muvaffaqiyatli — HTTP javobi qaytadi
	codeHash := sc.PhoneCodeHash

	twoFA := false // SignIn 2FA talab qilgani aniqlangach true bo'ladi
	for {
		var req verifyReq
		select {
		case req = <-pl.verifyCh:
		case <-time.After(loginTimeout):
			return errors.New("login vaqti tugadi")
		case <-ctx.Done():
			return ctx.Err()
		}

		// 2FA kerakligi allaqachon aniqlangan bo'lsa, qaytadan SignIn qilmaymiz
		// (kod allaqachon ishlatilgan) — to'g'ridan-to'g'ri parolni tekshiramiz.
		if twoFA {
			if req.password == "" {
				pl.resCh <- verifyRes{need2FA: true}
				continue
			}
			if _, err := client.Auth().Password(ctx, req.password); err != nil {
				pl.resCh <- verifyRes{err: err}
				continue
			}
			break
		}

		_, err := client.Auth().SignIn(ctx, account.Phone, req.code, codeHash)
		if errors.Is(err, auth.ErrPasswordAuthNeeded) {
			twoFA = true
			if req.password != "" {
				if _, err := client.Auth().Password(ctx, req.password); err != nil {
					pl.resCh <- verifyRes{err: err}
					continue
				}
				break
			}
			pl.resCh <- verifyRes{need2FA: true}
			continue
		} else if err != nil {
			pl.resCh <- verifyRes{err: err}
			continue
		}
		break // muvaffaqiyatli (2FA siz)
	}

	self, err := client.Self(ctx)
	if err != nil {
		pl.resCh <- verifyRes{err: err}
		return err
	}
	if err := repository.ActivateTelegramAccount(m.db, account.ID, self.ID, self.Username); err != nil {
		m.log.Error("account aktivlashtirishda xato", zap.Error(err))
	}
	m.removeLogin(account.ID)
	pl.resCh <- verifyRes{done: true}
	return nil
}

// startHumoBot HUMOcardbotni resolve qiladi, chat bo'sh bo'lsa /start yuboradi.
func (m *Manager) startHumoBot(ctx context.Context, client *telegram.Client, botID *atomic.Int64) {
	resolved, err := client.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: humoBotUsername})
	if err != nil {
		m.log.Warn("HUMOcardbot topilmadi", zap.Error(err))
		return
	}
	var bot *tg.User
	for _, u := range resolved.Users {
		if usr, ok := u.(*tg.User); ok && usr.Bot {
			bot = usr
			break
		}
	}
	if bot == nil {
		m.log.Warn("HUMOcardbot user sifatida topilmadi")
		return
	}
	botID.Store(bot.ID)

	peer := &tg.InputPeerUser{UserID: bot.ID, AccessHash: bot.AccessHash}
	// Chatda tarix bormi tekshiramiz — bo'sh bo'lsa /start bosamiz.
	if m.chatIsEmpty(ctx, client, peer) {
		sender := message.NewSender(client.API())
		if _, err := sender.To(peer).Text(ctx, "/start"); err != nil {
			m.log.Warn("/start yuborilmadi", zap.Error(err))
		} else {
			m.log.Info("HUMOcardbotga /start yuborildi")
		}
	}
}

func (m *Manager) chatIsEmpty(ctx context.Context, client *telegram.Client, peer tg.InputPeerClass) bool {
	hist, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{Peer: peer, Limit: 1})
	if err != nil {
		return false // shubha bo'lsa /start yubormaymiz (spam oldini olish)
	}
	switch h := hist.(type) {
	case *tg.MessagesMessages:
		return len(h.Messages) == 0
	case *tg.MessagesMessagesSlice:
		return len(h.Messages) == 0
	case *tg.MessagesChannelMessages:
		return len(h.Messages) == 0
	}
	return false
}

// onMessage HUMOcardbotdan kelgan to'lov xabarini qayta ishlaydi.
func (m *Manager) onMessage(ctx context.Context, account domain.TelegramAccount, botID int64, update *tg.UpdateNewMessage) error {
	msg, ok := update.Message.(*tg.Message)
	if !ok {
		return nil
	}
	text := msg.Message
	if text == "" {
		return nil
	}

	senderID := senderOf(msg)
	if botID != 0 && senderID != botID {
		return nil // faqat HUMOcardbotdan kelgan xabarlar
	}

	res := usecase.ParseTopUp(text, m.log)
	if res == nil {
		return nil
	}

	card, err := repository.GetCardByLast4(m.db, account.ID, res.Last4)
	if err != nil {
		m.log.Debug("Cartaga mos kelmadi", zap.String("last4", res.Last4))
		return nil
	}

	transID, err := repository.GetTransaction(m.db, card.ID, res.AmountInt, m.cfg.TimeoutMins)
	if err != nil {
		m.log.Info("Transaction topilmadi", zap.Int64("card", card.ID), zap.Int64("amount", res.AmountInt))
		return nil
	}

	m.log.Info("To'lov aniqlandi",
		zap.Int64("account", account.ID),
		zap.Int64("card", card.ID),
		zap.String("last4", res.Last4),
		zap.Int64("amount", res.AmountInt),
		zap.String("transaction_id", transID),
	)

	m.tasks <- domain.WebhookTask{
		UserID:  account.UserID,
		CardID:  card.ID,
		TransID: transID,
		Amount:  res.AmountInt,
		Action:  "confirm",
	}
	return nil
}

// senderOf shaxsiy chatda bo'sh FromID holatini PeerID orqali to'ldiradi.
func senderOf(msg *tg.Message) int64 {
	switch from := msg.FromID.(type) {
	case *tg.PeerUser:
		return from.UserID
	}
	if msg.Out {
		return 0
	}
	if peer, ok := msg.PeerID.(*tg.PeerUser); ok {
		return peer.UserID
	}
	return 0
}
