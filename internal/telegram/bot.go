package telegram

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type Bot struct {
	Bot *tele.Bot
	Log *zap.Logger

	mu      sync.Mutex
	started bool
}

func New(token string, apiURL string, clientTimeout time.Duration, log *zap.Logger) (*Bot, error) {
	tb, err := tele.NewBot(tele.Settings{
		Token: token,
		URL:   apiURL,
		Client: &http.Client{
			Timeout: clientTimeout,
		},
		Poller: &tele.LongPoller{
			Timeout: 10 * time.Second,
		},
		OnError: func(err error, _ tele.Context) {
			if IsHandledError(err) {
				return
			}
			log.Error("telegram polling error", zap.Error(err))
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return &Bot{
		Bot: tb,
		Log: log,
	}, nil
}

func (b *Bot) Start() {
	b.mu.Lock()
	if b.started {
		b.mu.Unlock()
		return
	}
	b.started = true
	b.mu.Unlock()

	b.Log.Info("telegram bot started", zap.String("bot_username", b.Bot.Me.Username))
	go b.Bot.Start()
}

func (b *Bot) Stop() {
	b.mu.Lock()
	if !b.started {
		b.mu.Unlock()
		return
	}
	b.started = false
	b.mu.Unlock()

	b.Bot.Stop()
	b.Log.Info("telegram bot stopped")
}
