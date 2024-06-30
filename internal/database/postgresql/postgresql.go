package postgresql

import (
	"context"
	"fmt"
	"grpc/internal/config"
	"grpc/internal/lib/database/repeateble"
	"grpc/internal/lib/logger/sl"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConection(ctx context.Context, log *slog.Logger, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	const op = "database.postgresql.NewClient"

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	var pool *pgxpool.Pool

	err := repeateble.DoWithTries(func() error {
		log.Info("database connection attempt")
		ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()

		pool, _ = pgxpool.New(ctx, dsn)
		err := pool.Ping(ctx)

		if err != nil {
			log.Error("database conection failed")
		}

		return err
	}, cfg.Attempts, cfg.Delay)

	if err != nil {
		log.Error("failed connect to database", sl.OpErr(op, err))
		return nil, err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		stopSignal := <-stop
		log.Info("stoppping database connection", slog.String("op", op), slog.String("signal", stopSignal.String()))
		pool.Close()
		log.Info("database was stopped")
	}()

	return pool, nil
}
