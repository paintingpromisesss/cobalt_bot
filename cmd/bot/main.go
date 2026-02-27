package main

import (
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/bot"
	"github.com/paintingpromisesss/cobalt_bot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("config error: %v", err))
	}
	if err := bot.Run(cfg); err != nil {
		panic(fmt.Sprintf("bot run error: %v", err))
	}
}
