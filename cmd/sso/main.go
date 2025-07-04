package main

import (
	"auth-service/internal/app"
	"auth-service/internal/config"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("Loading config")

	application := app.New(log, cfg.GRPC.Port, cfg.Database_url, cfg.TokenTTL)

	application.GRPCServer.MustRun()
	fmt.Println(cfg)

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	log.Info("Signal", sign.String())
	application.GRPCServer.Stop()
	log.Info("Shutting down...")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return log
}
