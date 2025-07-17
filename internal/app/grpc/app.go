package grpcApp

import (
	authgrpc "auth-service/internal/grpc/auth"
	"auth-service/internal/services/auth"
	"context"

	"log/slog"

	"net/http"
	"os"
	"time"
)

type App struct {
	log        *slog.Logger
	controller *http.Server
	port       string
}

func New(log *slog.Logger,
	authService *auth.Auth,
	port string) *App {

	controllers := authgrpc.Register(*authService, port)

	return &App{log, controllers, port}

}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}
func (a *App) Run() error {
	const op = "authapp.Run"

	log := a.log.With(slog.String("op", op), slog.String("port", a.port))

	go func() {
		if err := a.controller.ListenAndServe(); err != nil {
			log.Info("start", slog.String("running on", a.port))
		}
	}()
	return nil
}

func (a *App) Stop() {
	const op = "authApp.Stop"
	a.log.Info("shutting down")
	a.log.With(slog.String("op", op)).Info("stopping")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	a.controller.Shutdown(ctx)

	os.Exit(0)
}
