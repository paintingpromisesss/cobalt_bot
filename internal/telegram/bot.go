package telegram

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type Bot struct {
	Bot *tele.Bot
	Log *zap.Logger
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

func (b *Bot) Run(ctx context.Context) {
	b.Log.Info("telegram bot started", zap.String("bot_username", b.Bot.Me.Username))

	go b.Bot.Start()

	<-ctx.Done()
	b.Bot.Stop()
	b.Log.Info("telegram bot stopped")
}
