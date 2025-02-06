package delivery

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getUserID извлекает и валидирует user_id из контекста
func getUserID(c *gin.Context) (int, error) {
	userIDCtx, exists := c.Get("user_id")
	if !exists || userIDCtx == nil {
		return 0, fmt.Errorf("User ID missing in context")
	}

	userID, ok := userIDCtx.(int)
	if !ok {
		return 0, fmt.Errorf("Invalid user ID type in context")
	}

	return userID, nil
}

// validateUserID проверяет совпадание ID из контекста и из параметра запроса
func validateUserID(c *gin.Context) (int, bool) {
	userID, err := getUserID(c)
	if err != nil {
		logAndHandleError(c, http.StatusUnauthorized, err.Error(), err)
		return 0, false
	}

	userIDParam, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logAndHandleError(c, http.StatusBadRequest, "Invalid user ID type", err)
		return 0, false
	}

	if userID != userIDParam {
		logAndHandleError(c, http.StatusUnauthorized, "User ID mismatch", nil)
		return 0, false
	}

	return userID, true
}

// logAndHandleError логирует ошибку и отправляет HTTP-ответ
func logAndHandleError(c *gin.Context, status int, message string, err error) {
	if err != nil {
		slog.Error(message, "method", c.Request.Method, "path", c.Request.URL.Path, "client_ip", c.ClientIP(), "error", err)
	}
	c.JSON(status, gin.H{"error": message})
}
