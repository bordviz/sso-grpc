package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	jwt.RegisteredClaims
	UserID  int64         `json:"user_id"`
	AppID   int           `json:"app_id"`
	Expires time.Duration `json:"exp"`
}

type TokensPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
