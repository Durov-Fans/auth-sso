package app

import (
	"auth-service/internal/app/grpc"
	"auth-service/internal/services/auth"
	"auth-service/internal/storage/postgres"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcApp.App
}

func New(log *slog.Logger, grpcPort string, storageUrl string, tokenTTL time.Duration, tgToken string) *App {

	storage, err := postgres.InitDB(storageUrl)
	if err != nil {
		panic(err)
	}
	authService := auth.New(log, storage, storage, storage, tokenTTL, tgToken)

	grpcApp := grpcApp.New(log, authService, grpcPort)
	return &App{
		GRPCServer: grpcApp,
	}
}
