package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/handlers"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type Bot struct {
	bot *tele.Bot
	log *zap.Logger
}

func New(token string, log *zap.Logger) (*Bot, error) {
	tb, err := tele.NewBot(tele.Settings{
		Token: token,
		Poller: &tele.LongPoller{
			Timeout: 10 * time.Second,
		},
		OnError: func(err error, _ tele.Context) {
			log.Error("telegram polling error", zap.Error(err))
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	handlers.Register(tb)

	return &Bot{
		bot: tb,
		log: log,
	}, nil
}

func (b *Bot) Run(ctx context.Context) {
	b.log.Info("telegram bot started", zap.String("bot_username", b.bot.Me.Username))

	go b.bot.Start()

	<-ctx.Done()
	b.bot.Stop()
	b.log.Info("telegram bot stopped")
}
