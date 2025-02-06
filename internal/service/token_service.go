package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"user-management/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService interface {
	GenerateToken(ctx context.Context, userID int) (string, time.Time, error)
	ValidateToken(ctx context.Context, token string) (int, error)
	RevokeToken(ctx context.Context, token string) error
}
type DefaultTokenService struct {
	repo      repository.TokenRepository
	secretKey string
	logger    *slog.Logger
}

func NewTokenService(repo repository.TokenRepository, secretKey string, logger *slog.Logger) *DefaultTokenService {
	return &DefaultTokenService{
		repo:      repo,
		secretKey: secretKey,
		logger:    logger,
	}
}

// GenerateToken генерирует токен
func (s *DefaultTokenService) GenerateToken(ctx context.Context, userID int) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt.Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		s.logger.Error("Failed to generate token", "method", "GenerateToken", "user_id", userID, "error", err)
		return "", time.Time{}, err
	}

	err = s.repo.StoreToken(ctx, userID, tokenString, expiresAt)
	if err != nil {
		s.logger.Error("Failed to store token in database", "method", "GenerateToken", "user_id", userID, "error", err)
		return "", time.Time{}, err
	}
	s.logger.Info("Token successfully generated and stored", "method", "GenerateToken", "user_id", userID, "expires_at", expiresAt)
	return tokenString, expiresAt, nil
}

// ValidateToken проверяет валидность токена
func (s *DefaultTokenService) ValidateToken(ctx context.Context, tokenString string) (int, error) {
	parsedToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Проверка метода подписи токена
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			s.logger.Error("Invalid token signing method", "method", "ValidateToken", "error", err, "token", tokenString)
			return nil, err
		}
		return []byte(s.secretKey), nil
	})
	if err != nil || !parsedToken.Valid {
		s.logger.Error("Invalid token", "method", "ValidateToken", "token", tokenString, "error", err)
		return 0, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		s.logger.Error("Invalid token claims", "method", "ValidateToken", "token", tokenString)
		return 0, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		s.logger.Error("Invalid user ID in token claims", "method", "ValidateToken", "token", tokenString)
		return 0, errors.New("invalid user ID in token")
	}

	isValid, err := s.repo.IsTokenValid(ctx, tokenString)
	if err != nil {
		s.logger.Error("Failed to validate token in database", "method", "ValidateToken", "token", tokenString, "error", err)
	}

	if !isValid {
		s.logger.Warn("Token is invalid or revoked", "method", "ValidateToken", "token", tokenString)
		return 0, errors.New("token is invalid or revoked")
	}

	s.logger.Info("Token validated successfully", "method", "ValidateToken", "user_id", int(userID))
	return int(userID), nil
}

// RevokeToken отзывает токен
func (s *DefaultTokenService) RevokeToken(ctx context.Context, token string) error {
	err := s.repo.RevokeToken(ctx, token)
	if err != nil {
		s.logger.Error("Failed to revoke token", "method", "RevokeToken", "token", token, "error", err)
		return err
	}

	s.logger.Info("Token successfully revoked", "method", "RevokeToken", "token", token)
	return nil
}
