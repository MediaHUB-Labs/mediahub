package auth

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID uint, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	// Use JWT_EXPIRY_MINUTES from env, fallback to 1440 (24 hours)
	expiryMinutes := 1440
	if envExpiry := os.Getenv("JWT_EXPIRY_MINUTES"); envExpiry != "" {
		if parsed, err := strconv.Atoi(envExpiry); err == nil && parsed > 0 {
			expiryMinutes = parsed
		}
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Duration(expiryMinutes) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenStr string) (*jwt.Token, error) {

	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}
