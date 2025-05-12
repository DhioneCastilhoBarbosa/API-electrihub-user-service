package middlewares

import (
	"net/http"
	"strings"
	"user-service/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const bearerPrefix = "Bearer "

// AuthMiddleware protege rotas que exigem autenticação JWT.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Verifica se o Authorization header está presente e formatado corretamente
		if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token ausente ou mal formatado (esperado: Bearer <token>)",
			})
			c.Abort()
			return
		}

		// Extrai o token do header
		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		// Prepara estrutura para os claims do token
		claims := &utils.Claims{}

		// Valida e parseia o token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return utils.SecretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token inválido ou expirado",
			})
			c.Abort()
			return
		}

		// Token válido — injeta dados no contexto da requisição
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
