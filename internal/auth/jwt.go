package auth

import (
	"errors"
	"fmt"

	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/golang-jwt/jwt/v4"
)

type claims struct {
	jwt.RegisteredClaims
	ID uint
}

// BuildJWTString - creates and return a string representation of JWT token
func BuildJWTString(id uint) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims{
			RegisteredClaims: jwt.RegisteredClaims{},
			ID:               id,
		},
	)

	tokenString, err := token.SignedString([]byte(config.Config.AuthSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserIDFromToken(tokenString string) (uint, error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(config.Config.AuthSecretKey), nil
		},
	)

	if err != nil {
		return 0, fmt.Errorf("error occured while parsing JWT token: %w", err)
	}

	if !token.Valid {
		logger.Log.Warn("Token is not valid")
		return 0, errors.New("token is not valid")
	}

	logger.Log.Info("Token is valid")
	return claims.ID, nil
}
