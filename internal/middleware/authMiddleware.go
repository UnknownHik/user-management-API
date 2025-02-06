package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"user-management/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenService service.TokenService
	logger       *slog.Logger
}

func NewAuthMiddleware(tokenService service.TokenService, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		logger:       logger,
	}
}

// AuthMiddleware проверяет JWT токен и добавляет `user_id` в GIN контекст
func (m *AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := m.tokenService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
