package main

import (
	"auth-service/internal/config"
	"fmt"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
)

func main() {
	cfg := config.MustLoad()
	log := SetupLogger(cfg.Env)

	log.Info("Loading config")
	fmt.Println(cfg)

}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
