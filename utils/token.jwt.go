package utils

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// GenerateToken membuat token JWT baru untuk user tertentu
func GenerateToken(userId uint) (string, error) {
	godotenv.Load()
	secret := os.Getenv("JWT_SECRET")

	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ReverseToken memverifikasi token dan mengembalikan user_id
func ReverseToken(tokenStr string) (uint, error) {
	godotenv.Load()
	secret := os.Getenv("JWT_SECRET")

	if tokenStr == "" {
		return 0, errors.New("token tidak boleh kosong")
	}

	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return 0, errors.New("user_id tidak valid di token")
		}
		return uint(userIDFloat), nil
	}

	return 0, errors.New("token tidak valid")
}
