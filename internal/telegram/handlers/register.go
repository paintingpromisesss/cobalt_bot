package handlers

import tele "gopkg.in/telebot.v4"

func Register(tb *tele.Bot) {
	tb.Handle("/start", func(c tele.Context) error {
		return c.Send("Бот запущен.")
	})
}
