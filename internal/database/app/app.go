package app

import (
	"context"
	"errors"
	"grpc/internal/domain/models"
	"grpc/internal/lib/database/query"
	"grpc/internal/lib/logger/sl"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppDB struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewAppDB(pool *pgxpool.Pool, log *slog.Logger) *AppDB {
	return &AppDB{
		pool: pool,
		log:  log,
	}
}

func (a *AppDB) GetAppByID(ctx context.Context, appID int) (models.App, error) {
	const op = "database.app.GetAppByID"

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		a.log.Error("failed to begin transaction", sl.OpErr(op, err))
		return models.App{}, err
	}
	defer tx.Rollback(ctx)

	q := `
		SELECT id, name, secret, refresh_secret FROM app WHERE id = $1;
	`

	a.log.Debug("get app by id query", slog.String("op", op), slog.String("query", query.QueryToString(q)))

	var app models.App

	err = tx.QueryRow(ctx, q, appID).Scan(&app.ID, &app.Name, &app.Secret, &app.RefreshSecret)
	if err != nil {
		if err == pgx.ErrNoRows {
			a.log.Error("app not found", slog.String("op", op), slog.Int("id", appID))
			return models.App{}, errors.New("app not found")
		}
		a.log.Error("failed to get app by id", sl.OpErr(op, err))
		return models.App{}, err
	}

	a.log.Info("successfully get app by id", slog.String("op", op), slog.Int("id", appID))
	return app, nil
}
