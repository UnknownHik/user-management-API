package delivery

import (
	"log/slog"
	"net/http"
	"strings"

	"user-management/internal/config"
	"user-management/internal/dto"
	"user-management/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService  service.UserService
	tokenService service.TokenService
	logger       *slog.Logger
	config       *config.Config
}

func NewUserHandler(userService service.UserService, tokenService service.TokenService, config *config.Config, logger *slog.Logger) UserHandler {
	return UserHandler{
		userService:  userService,
		tokenService: tokenService,
		logger:       logger,
		config:       config,
	}
}

// RegisterHandler обрабатывает запрос на регистрацию пользователя
func (h *UserHandler) RegisterHandler(c *gin.Context) {
	var userDTO dto.UserRegLogDTO

	if err := c.ShouldBindJSON(&userDTO); err != nil {
		logAndHandleError(c, http.StatusBadRequest, "Invalid username or password", err)
		return
	}

	userID, err := h.userService.Register(c.Request.Context(), &userDTO)
	if err != nil {
		logAndHandleError(c, http.StatusInternalServerError, "Error during user registration", err)
		return
	}

	h.logger.Info("User registered successfully", "method", "RegisterHandler", "username", userDTO.UserName, "user_id", userID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "Успешная регистрация",
		"user_id": userID,
	})
}

// LoginHandler обрабатывает запрос на авторизацию пользователя
func (h *UserHandler) LoginHandler(c *gin.Context) {
	var userDTO dto.UserRegLogDTO

	if err := c.ShouldBindJSON(&userDTO); err != nil {
		logAndHandleError(c, http.StatusBadRequest, "Error binding login input", err)
		return
	}

	user, err := h.userService.Login(c.Request.Context(), &userDTO)
	if err != nil {
		logAndHandleError(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	token, expiresAt, err := h.tokenService.GenerateToken(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	h.logger.Info("User logged in successfully", "method", "LoginHandler", "username", user.UserName, "user_id", user.ID)
	c.JSON(http.StatusOK, gin.H{
		"status":     "Авторизация успешна",
		"token":      token,
		"expires_at": expiresAt,
	})
}

// LogoutHandler обрабатывает выход пользователя из системы
func (h *UserHandler) LogoutHandler(c *gin.Context) {
	token := c.GetHeader("Authorization")

	token = strings.TrimPrefix(token, "Bearer ")
	err := h.tokenService.RevokeToken(c.Request.Context(), token)
	if err != nil {
		logAndHandleError(c, http.StatusInternalServerError, "Failed to revoke token", err)
		return
	}

	h.logger.Info("Token revoked successfully", "method", "LogoutHandler", "token", token)
	c.JSON(http.StatusOK, gin.H{
		"status": "Вышел из системы",
	})
}

// UserStatusHandler обрабатывает запрос на получение информации о пользователе
func (h *UserHandler) UserStatusHandler(c *gin.Context) {
	userID, ok := validateUserID(c)
	if !ok {
		return
	}

	userStatus, err := h.userService.UserStatus(c.Request.Context(), userID)
	if err != nil {
		logAndHandleError(c, http.StatusInternalServerError, "Failed get user status", err)
		return
	}

	h.logger.Info("User status return successfully", "method", "UserStatusHandler", "user_id", userID)
	c.JSON(http.StatusOK, userStatus)
}

// UsersLeaderboardHandler обрабатывает запрос на получение списка топ пользователей с самым большим балансом
func (h *UserHandler) UsersLeaderboardHandler(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		logAndHandleError(c, http.StatusUnauthorized, err.Error(), err)
		return
	}

	leaderboard, err := h.userService.UserLeaderboard(c.Request.Context())
	if err != nil {
		logAndHandleError(c, http.StatusInternalServerError, "Failed get leaderboard", err)
		return
	}

	h.logger.Info("Leaderboard return successfully", "method", "UsersLeaderboardHandler", "user_id", userID)
	c.JSON(http.StatusOK, leaderboard)
}

// ReferrerHandler обрабатывает запрос на добавление реферального кода
func (h *UserHandler) ReferrerHandler(c *gin.Context) {
	userID, ok := validateUserID(c)
	if !ok {
		return
	}

	var referrer dto.ReferrerDTO

	if err := c.ShouldBindJSON(&referrer); err != nil {
		logAndHandleError(c, http.StatusBadRequest, "Error binding referrer", err)
		return
	}

	err := h.userService.AddReferrer(c.Request.Context(), userID, &referrer)
	if err != nil {
		logAndHandleError(c, http.StatusInternalServerError, "Error adding referrer", err)
		return
	}

	h.logger.Info("Referrer added successfully", "method", "ReferralHandler", "userID", userID, "referrer", referrer.Referrer)
	c.JSON(http.StatusOK, gin.H{
		"status":      "Реферер добавлен",
		"referrer_id": referrer.Referrer,
	})
}

// TaskCompleteHandler обрабатывает запрос на выполнение задания
func (h *UserHandler) TaskCompleteHandler(c *gin.Context) {
	userID, ok := validateUserID(c)
	if !ok {
		return
	}

	var task dto.TaskDTO

	if err := c.ShouldBindJSON(&task); err != nil {
		logAndHandleError(c, http.StatusBadRequest, "Error binding task", err)
		return
	}

	err := h.userService.TaskComplete(c.Request.Context(), userID, &task)
	if err != nil {
		logAndHandleError(c, http.StatusInternalServerError, "Error completing task", err)
		return
	}

	h.logger.Info("Task completed successfully", "method", "TaskComplete", "userID", userID, "task", task.ID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "Задание выполнено",
		"task_id": task.ID,
	})
}
