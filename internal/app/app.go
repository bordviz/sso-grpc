package app

import (
	"context"
	grpcapp "grpc/internal/app/grpc"
	"grpc/internal/config"
	appdb "grpc/internal/database/app"
	authdb "grpc/internal/database/auth"
	"grpc/internal/database/postgresql"
	"grpc/internal/lib/logger/sl"
	authservice "grpc/internal/services/auth"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	const op = "app.New"

	dbPool, err := postgresql.NewConection(context.TODO(), log, cfg.Database)
	if err != nil {
		log.Error("failed connect to database", sl.OpErr(op, err))
	}

	authDB := authdb.NewAuthDB(dbPool, log)
	appDB := appdb.NewAppDB(dbPool, log)
	db := authservice.DB{
		AuthDB: authDB,
		AppDB:  appDB,
	}

	authService := authservice.NewAuthService(log, db, cfg.TokenExpires, cfg.RefreshTokenExpires)

	grpcApp := grpcapp.New(log, cfg.GRPC.Port, authService)

	return &App{
		GRPCSrv: grpcApp,
	}
}
