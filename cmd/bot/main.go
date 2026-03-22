package main

import (
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/bootstrap"
	"github.com/paintingpromisesss/cobalt_bot/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("config error: %v", err))
	}
	if err := bootstrap.Run(cfg); err != nil {
		panic(fmt.Sprintf("app run error: %v", err))
	}
}
