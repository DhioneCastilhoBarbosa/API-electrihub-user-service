package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Chave secreta para assinar o token
var SecretKey = []byte(os.Getenv("JWT_SECRET"))

// Claims personalizados para o JWT
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Gera um token JWT para o usuário
func GenerateJWT(userID string, role string) (string, error) {
	expirationTime := time.Now().Add(time.Hour)

	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}
