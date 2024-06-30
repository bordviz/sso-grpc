package jwt

import (
	"fmt"
	"grpc/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(userID int64, appID int, secret string, expires time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)

	fmt.Println(userID, appID, secret, expires)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["app_id"] = appID
	claims["exp"] = time.Now().Add(expires).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func DecodeToken(token string, secret string) (models.Token, error) {
	var model models.Token

	jwtToken, err := jwt.ParseWithClaims(token, &model, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !jwtToken.Valid {
		return models.Token{}, fmt.Errorf("failed to decode token: %w", err)
	}

	return model, nil
}
