package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateRefreshToken() (string, error) {
	// Simple random string for refresh token, could be more robust
	return GenerateRandomString(32)
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}
