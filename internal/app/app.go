package app

import (
	"auth-service/internal/app/grpc"
	"google.golang.org/grpc"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcApp.App
}

func New(log *slog.Logger, grpcPort int, storageUrl string, tokenTTL time.Duration) *App {

	grpcApp := grpcApp.New(log, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
