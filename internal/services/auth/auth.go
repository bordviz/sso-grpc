package auth

import (
	"context"
	"errors"
	"grpc/internal/domain/models"
	"grpc/internal/lib/jwt"
	"grpc/internal/lib/logger/sl"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthDB interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByID(ctx context.Context, userID int64) (models.User, error)
	IsAdmin(ctx context.Context, userID int64, appID int) (bool, error)
	CheckUser(ctx context.Context, email string) (bool, error)
}

type AppDB interface {
	GetAppByID(ctx context.Context, appID int) (models.App, error)
}

type DB struct {
	AuthDB AuthDB
	AppDB  AppDB
}

type AuthService struct {
	log                 *slog.Logger
	db                  DB
	tokenExpires        time.Duration
	refreshTokenExpires time.Duration
}

var (
	ErrInvPassOrEmail   = errors.New("invalid password or email")
	ErrUserAlreadyExist = errors.New("user already registered")
	ErrInvalidData      = errors.New("invalid data")
	ErrUnauthorized     = errors.New("unauthorized")
)

func NewAuthService(log *slog.Logger, db DB, tokenExpires time.Duration, refreshTokenExpires time.Duration) *AuthService {
	return &AuthService{
		log:                 log,
		db:                  db,
		tokenExpires:        tokenExpires,
		refreshTokenExpires: refreshTokenExpires,
	}
}

func (a *AuthService) Register(ctx context.Context, email string, password string, name string) (int64, error) {
	const op = "services.auth.Register"

	checkUser, err := a.db.AuthDB.CheckUser(ctx, email)
	if err != nil {
		a.log.Error("failed to check user", sl.OpErr(op, err))
		return 0, err
	}
	if checkUser {
		a.log.Info("user already exists", slog.String("op", op), slog.String("email", email))
		return 0, ErrUserAlreadyExist
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed generate hash password", sl.OpErr(op, err))
		return 0, err
	}

	user := models.User{
		Email:    email,
		PassHash: string(hashPassword),
		Name:     name,
	}

	userID, err := a.db.AuthDB.CreateUser(ctx, user)
	if err != nil {
		a.log.Error("failed to save user on database", sl.OpErr(op, err))
		return 0, err
	}

	a.log.Info("new user created", slog.String("op", op), slog.Int64("id", userID))
	return userID, nil
}

func (a *AuthService) Login(ctx context.Context, email string, password string, appID int) (models.TokensPair, error) {
	const op = "services.auth.Login"

	user, err := a.db.AuthDB.GetUserByEmail(ctx, email)
	if err != nil {
		a.log.Error("failed to get user by email", sl.OpErr(op, err))
		return models.TokensPair{}, ErrInvPassOrEmail
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password)); err != nil {
		a.log.Error("invalid password", sl.OpErr(op, err))
		return models.TokensPair{}, ErrInvPassOrEmail
	}

	app, err := a.db.AppDB.GetAppByID(ctx, appID)
	if err != nil {
		a.log.Error("failed to get app by id", sl.OpErr(op, err))
		return models.TokensPair{}, ErrInvalidData
	}

	if app.Secret == "" || app.RefreshSecret == "" {
		a.log.Error("app secret is empty", slog.String("op", op))
		return models.TokensPair{}, ErrInvalidData
	}

	token, err := jwt.CreateToken(user.ID, app.ID, app.Secret, a.tokenExpires)
	if err != nil {
		a.log.Error("failed to create token", sl.OpErr(op, err))
		return models.TokensPair{}, err
	}
	refreshToken, err := jwt.CreateToken(user.ID, app.ID, app.RefreshSecret, a.refreshTokenExpires)
	if err != nil {
		a.log.Error("failed to create token", sl.OpErr(op, err))
		return models.TokensPair{}, err
	}

	tokensPair := models.TokensPair{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}

	a.log.Info("user login complete", slog.String("op", op), slog.String("email", email))
	return tokensPair, nil

}

func (a *AuthService) IsAdmin(ctx context.Context, userID int64, appID int) (bool, error) {
	const op = "services.auth.IsAdmin"

	isAdmin, err := a.db.AuthDB.IsAdmin(ctx, userID, appID)
	if err != nil {
		a.log.Error("failed to check user is admin", sl.OpErr(op, err))
		return false, err
	}

	a.log.Info("user is admin",
		slog.String("op", op),
		slog.Int64("id", userID),
		slog.Int("app_id", appID),
		slog.Bool("is_admin", isAdmin),
	)
	return isAdmin, nil
}

func (a *AuthService) RefreshToken(ctx context.Context, token string, appID int) (tokens models.TokensPair, err error) {
	const op = "services.auth.RefreshToken"

	app, err := a.db.AppDB.GetAppByID(ctx, appID)
	if err != nil {
		a.log.Error("failed to get app by id", sl.OpErr(op, err))
		return models.TokensPair{}, ErrInvalidData
	}

	if app.Secret == "" || app.RefreshSecret == "" {
		a.log.Error("app secret is empty", slog.String("op", op))
		return models.TokensPair{}, ErrInvalidData
	}

	decodeToken, err := jwt.DecodeToken(token, app.RefreshSecret)
	if err != nil {
		a.log.Error("failed to decode token", sl.OpErr(op, err))
		return models.TokensPair{}, ErrUnauthorized
	}

	accessToken, err := jwt.CreateToken(decodeToken.UserID, app.ID, app.Secret, a.tokenExpires)
	if err != nil {
		a.log.Error("failed to create token", sl.OpErr(op, err))
		return models.TokensPair{}, err
	}
	refreshToken, err := jwt.CreateToken(decodeToken.UserID, app.ID, app.RefreshSecret, a.refreshTokenExpires)
	if err != nil {
		a.log.Error("failed to create token", sl.OpErr(op, err))
		return models.TokensPair{}, err
	}

	tokensPair := models.TokensPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	a.log.Info("user refresh token complete", slog.String("op", op), slog.Int64("id", decodeToken.UserID))
	return tokensPair, nil
}

func (a *AuthService) CurrentUser(ctx context.Context, token string, appID int) (models.UserRead, error) {
	const op = "services.auth.CurrentUser"

	app, err := a.db.AppDB.GetAppByID(ctx, appID)
	if err != nil {
		a.log.Error("failed to get app by id", sl.OpErr(op, err))
		return models.UserRead{}, ErrInvalidData
	}

	if app.Secret == "" || app.RefreshSecret == "" {
		a.log.Error("app secret is empty", slog.String("op", op))
		return models.UserRead{}, ErrInvalidData
	}

	decodeToken, err := jwt.DecodeToken(token, app.Secret)
	if err != nil {
		a.log.Error("failed to decode token", sl.OpErr(op, err))
		return models.UserRead{}, ErrUnauthorized
	}

	user, err := a.db.AuthDB.GetUserByID(ctx, decodeToken.UserID)
	if err != nil {
		a.log.Error("failed to get user by id", sl.OpErr(op, err))
		return models.UserRead{}, ErrInvalidData
	}

	a.log.Info("get user complete", slog.String("op", op), slog.Int64("id", decodeToken.UserID))
	return models.UserRead{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}
