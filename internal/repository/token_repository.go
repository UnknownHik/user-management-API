package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type TokenRepository interface {
	StoreToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
	IsTokenValid(ctx context.Context, token string) (bool, error)
	RevokeToken(ctx context.Context, token string) error
}

type TokenRepo struct {
	db     *pgx.Conn
	logger *slog.Logger
}

func NewTokenRepo(db *pgx.Conn, logger *slog.Logger) *TokenRepo {
	return &TokenRepo{
		db:     db,
		logger: logger,
	}
}

// SQL запросы
const (
	queryStoreToken        = `INSERT INTO tokens(user_id, token, expires_at, is_revoked) VALUES ($1, $2, $3, FALSE)`
	queryIsTokenValid      = `SELECT EXISTS (SELECT 1 FROM tokens WHERE token = $1 AND expires_at > NOW() AND is_revoked = FALSE)`
	queryUpdateTokenRevoke = `UPDATE tokens SET is_revoked = TRUE WHERE token = $1`
)

// StoreToken сохраняет токен в базе данных
func (tr *TokenRepo) StoreToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	tr.logger.Info("Executing query", "method", "StoreToken", "query", queryStoreToken, "token", token, "user_id", userID, "expires_at", expiresAt)

	result, err := tr.db.Exec(ctx, queryStoreToken, userID, token, expiresAt)
	if err != nil {
		return tr.handleError("StoreToken", "Failed to execute query to store token", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		tr.logger.Warn("No rows affected, token may already be revoked", "method", "StoreToken", "token", token)
		return fmt.Errorf("token not found or already revoked")
	}

	tr.logger.Info("Token successfully stored", "token", token)
	return nil
}

// IsTokenValid проверяет валидность токена
func (tr *TokenRepo) IsTokenValid(ctx context.Context, token string) (bool, error) {
	tr.logger.Info("Executing query", "method", "IsTokenValid", "query", queryIsTokenValid, "token", token)

	var isValid bool
	err := tr.db.QueryRow(ctx, queryIsTokenValid, token).Scan(&isValid)
	if err != nil {
		return false, tr.handleError("IsTokenValid", "Failed to execute query to validate token", err)
	}

	tr.logger.Info("Token validation result", "isValid", isValid, "token", token)
	return isValid, nil
}

// RevokeToken отзывает токен
func (tr *TokenRepo) RevokeToken(ctx context.Context, token string) error {
	tr.logger.Info("Executing query", "method", "RevokeToken", "query", queryUpdateTokenRevoke, "token", token)

	_, err := tr.db.Exec(ctx, queryUpdateTokenRevoke, token)
	if err != nil {
		return tr.handleError("RevokeToken", "Failed to execute query to revoke token", err)
	}
	tr.logger.Info("Token successfully revoked", "token", token)
	return nil
}

// handleError служит для обработки ошибок и логирования
func (tr *TokenRepo) handleError(method, message string, err error) error {
	tr.logger.Error("Error", "method", method, "error", err)
	return fmt.Errorf("%s: %w", message, err)
}
