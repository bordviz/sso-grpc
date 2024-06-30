package auth

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

type AuthDB struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewAuthDB(pool *pgxpool.Pool, log *slog.Logger) *AuthDB {
	return &AuthDB{
		pool: pool,
		log:  log,
	}
}

func (a *AuthDB) CreateUser(ctx context.Context, user models.User) (int64, error) {
	const op = "database.auth.CreateUser"

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		a.log.Error("failed to begin transaction", sl.OpErr(op, err))
		return 0, err
	}
	defer tx.Rollback(ctx)

	q := `
		INSERT INTO public.user (email, hash_password, name)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	a.log.Debug("create user query", slog.String("op", op), slog.String("query", query.QueryToString(q)))

	var id int64
	if err := tx.QueryRow(ctx, q, user.Email, user.PassHash, user.Name).Scan(&id); err != nil {
		a.log.Error("failed to create user", sl.OpErr(op, err))
		tx.Rollback(ctx)
		return 0, errors.New("failed to create user")
	}

	err = tx.Commit(ctx)
	if err != nil {
		a.log.Error("failed to commit transaction", sl.OpErr(op, err))
		return 0, err
	}

	a.log.Info("new user created", slog.String("op", op), slog.Int64("id", id))
	return id, nil
}

func (a *AuthDB) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "database.auth.GetUserByEmail"

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		a.log.Error("failed to begin transaction", sl.OpErr(op, err))
		return models.User{}, err
	}
	defer tx.Rollback(ctx)

	q := `
		SELECT id, name, email, hash_password FROM public.user WHERE email = $1;
	`

	a.log.Debug("get user by email query", slog.String("op", op), slog.String("query", query.QueryToString(q)))

	var user models.User

	err = tx.QueryRow(ctx, q, email).Scan(&user.ID, &user.Name, &user.Email, &user.PassHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			a.log.Error("user not found", slog.String("op", op), slog.String("email", email))
			return models.User{}, errors.New("user not found")
		}
		a.log.Error("failed to get user by email", sl.OpErr(op, err))
		return models.User{}, err
	}

	a.log.Info("successfully get user by email", slog.String("op", op), slog.String("email", email))
	return user, nil
}

func (a *AuthDB) GetUserByID(ctx context.Context, userID int64) (models.User, error) {
	const op = "database.auth.GetUserByID"

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		a.log.Error("failed to begin transaction", sl.OpErr(op, err))
		return models.User{}, err
	}
	defer tx.Rollback(ctx)

	q := `
		SELECT id, name, email, hash_password FROM public.user WHERE id = $1;
	`

	a.log.Debug("get user by id query", slog.String("op", op), slog.String("query", query.QueryToString(q)))

	var user models.User

	err = tx.QueryRow(ctx, q, userID).Scan(&user.ID, &user.Name, &user.Email, &user.PassHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			a.log.Error("user not found", slog.String("op", op), slog.Int64("id", userID))
			return models.User{}, errors.New("user not found")
		}
		a.log.Error("failed to get user by id", sl.OpErr(op, err))
		return models.User{}, err
	}

	a.log.Info("successfully get user by id", slog.String("op", op), slog.Int64("id", userID))
	return user, nil
}

func (a *AuthDB) IsAdmin(ctx context.Context, userID int64, appID int) (bool, error) {
	const op = "database.auth.IsAdmin"

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		a.log.Error("failed to begin transaction", sl.OpErr(op, err))
		return false, err
	}
	defer tx.Rollback(ctx)

	q := `
        SELECT id, user_id, app_id FROM admin 
		WHERE user_id = $1 AND app_id = $2;
    `

	a.log.Debug("is admin query", slog.String("op", op), slog.String("query", query.QueryToString(q)))

	var isAdmin models.Admin

	if err = tx.QueryRow(ctx, q, userID, appID).Scan(&isAdmin.ID, &isAdmin.UserID, &isAdmin.AppID); err != nil {
		if err == pgx.ErrNoRows {
			a.log.Error("user not found", slog.String("op", op), slog.Int64("id", userID))
			return false, nil
		}
		a.log.Error("failed to get user by id", sl.OpErr(op, err))
		return false, err
	}

	a.log.Info("successfully check user is admin", slog.String("op", op), slog.Int64("id", userID), slog.Bool("is_admin", true))
	return true, nil
}

func (a *AuthDB) CheckUser(ctx context.Context, email string) (bool, error) {
	const op = "database.auth.CheckUser"

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		a.log.Error("failed to begin transaction", sl.OpErr(op, err))
		return false, err
	}
	defer tx.Rollback(ctx)

	q := `
        SELECT id FROM public.user WHERE email = $1;
    `

	a.log.Debug("check user query", slog.String("op", op), slog.String("query", query.QueryToString(q)))

	var userID int

	if err := tx.QueryRow(ctx, q, email).Scan(&userID); err != nil {
		if err == pgx.ErrNoRows {
			a.log.Info("user not found", slog.String("op", op), slog.String("email", email))
			return false, nil
		}
		a.log.Error("failed to check user", sl.OpErr(op, err))
		return false, err
	}

	a.log.Info("successfully check user", slog.String("op", op), slog.String("email", email), slog.Int("id", userID))
	return true, nil
}
