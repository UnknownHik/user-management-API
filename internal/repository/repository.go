package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"user-management/internal/dto"
	"user-management/internal/models"

	"github.com/jackc/pgx/v5"
)

var (
	ErrFailedBeginTx      = errors.New("failed to begin transaction")
	ErrFailedExecuteQuery = errors.New("failed to execute query")
	ErrUserNotFound       = errors.New("user not found")
	ErrTaskNotFound       = errors.New("task not found")
)

type UserRepository interface {
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
	GetUserByNameWithTx(ctx context.Context, tx pgx.Tx, name string) (*models.User, error)
	CreateUserWithTx(ctx context.Context, tx pgx.Tx, user *models.User) (int, error)
	GetUserByName(ctx context.Context, name string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserLeaderboard(ctx context.Context) ([]dto.UserLeaderDTO, error)
	SetReferrer(ctx context.Context, tx pgx.Tx, userID, referrer int) error
	GetUserByIDWithTx(ctx context.Context, tx pgx.Tx, id int) (*models.User, error)
	AddPoint(ctx context.Context, tx pgx.Tx, userID int, points int) error
	GetTask(ctx context.Context, taskID int) (*models.Task, error)
	IsCompletedTask(ctx context.Context, tx pgx.Tx, userID, taskID int) (bool, error)
	AddCompletedTask(ctx context.Context, tx pgx.Tx, userID, taskID int) error
}

type UserRepo struct {
	db     *pgx.Conn
	logger *slog.Logger
}

func NewUserRepository(db *pgx.Conn, logger *slog.Logger) *UserRepo {
	return &UserRepo{db: db, logger: logger}
}

// SQL запросы
const (
	queryCreateUser             = `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
	queryGetUserByNameForUpdate = `SELECT id, username FROM users WHERE username = $1 FOR UPDATE`
	queryGetUserByName          = `SELECT id, username, password FROM users WHERE username = $1`
	queryGetUserByID            = `SELECT id, username, password, balance, updated_balance, referrer, created_at FROM users WHERE id = $1`
	queryGetUserByIDForUpdate   = `SELECT id, username, password, balance, referrer FROM users WHERE id = $1 FOR UPDATE`
	queryGetLeaderboard         = `SELECT id, username, balance FROM users ORDER BY balance DESC LIMIT 10`
	queryUpdateReferrer         = `UPDATE users SET referrer = $1 WHERE id = $2`
	queryUpdatePoints           = `UPDATE users SET balance = balance + $1, updated_balance = NOW() WHERE id = $2`
	queryGetTask                = `SELECT id, description, reward FROM tasks WHERE id = $1`
	queryIsCompletedTask        = `SELECT EXISTS (SELECT 1 FROM completed_tasks WHERE user_id = $1 AND task_id = $2)`
	queryCompletedTask          = `INSERT INTO completed_tasks (user_id, task_id) VALUES ($1, $2)`
)

// BeginTransaction начало транзакции
func (r *UserRepo) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.Error("Failed to begin transaction", "error", err)
		return nil, fmt.Errorf("BeginTransaction: %w", ErrFailedBeginTx)
	}
	r.logger.Info("Transaction started")
	return tx, nil
}

// GetUserByNameWithTx получение пользователя по имени с блокировкой для не атомарных операций
func (r *UserRepo) GetUserByNameWithTx(ctx context.Context, tx pgx.Tx, name string) (*models.User, error) {
	var user models.User

	r.logger.Info("Executing query", "query", queryGetUserByNameForUpdate, "username", name)
	err := tx.QueryRow(ctx, queryGetUserByNameForUpdate, name).Scan(&user.ID, &user.UserName)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Info("User not found", "username", name)
			return nil, fmt.Errorf("GetUserByNameWithTx: user not found: %w", err)
		}
		r.logger.Error("Failed to execute query to get user by name", "error", err, "username", name)
		return nil, fmt.Errorf("GetUserByName:  %w", ErrFailedExecuteQuery)
	}

	r.logger.Info("User found", "user_id", user.ID)
	return &user, nil
}

// CreateUserWithTx создание нового пользователя
func (r *UserRepo) CreateUserWithTx(ctx context.Context, tx pgx.Tx, user *models.User) (int, error) {
	var userID int

	r.logger.Info("Executing query", "query", queryCreateUser, "username", user.UserName)

	err := tx.QueryRow(ctx, queryCreateUser, user.UserName, user.Password).Scan(&userID)
	if err != nil {
		r.logger.Error("Failed to execute query create user", "error", err, "username", user.UserName)
		return 0, fmt.Errorf("CreateUser:  %w", ErrFailedExecuteQuery)
	}

	r.logger.Info("User created", "user_id", userID)
	return userID, nil
}

// GetUserByIDWithTx получение пользователя по id с блокировкой для не атомарных операций
func (r *UserRepo) GetUserByIDWithTx(ctx context.Context, tx pgx.Tx, id int) (*models.User, error) {
	var user models.User

	r.logger.Info("Executing query", "query", queryGetUserByIDForUpdate, "user_id", id)
	err := tx.QueryRow(ctx, queryGetUserByIDForUpdate, id).Scan(&user.ID, &user.UserName, &user.Password, &user.Balance, &user.Referrer)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Info("User not found", "user_id", id)
			return nil, fmt.Errorf("GetUserByIDWithTx: user not found: %w", err)
		}
		r.logger.Error("Failed to get user with FOR UPDATE", "error", err, "user_id", id)
		return nil, fmt.Errorf("GetUserByIDWithTx: %w", ErrFailedExecuteQuery)
	}

	r.logger.Info("User found", "user_id", user.ID)
	return &user, nil
}

// GetUserByName получение данных пользователя по имени для входа
func (r *UserRepo) GetUserByName(ctx context.Context, name string) (*models.User, error) {
	var user models.User

	r.logger.Info("Executing query", "query", queryGetUserByName, "username", name)
	err := r.db.QueryRow(ctx, queryGetUserByName, name).Scan(&user.ID, &user.UserName, &user.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Info("User not found", "username", name)
			return nil, fmt.Errorf("GetUserByName:  %w", ErrUserNotFound)
		}

		r.logger.Error("Failed to execute query to get user by name", "error", err, "username", name)
		return nil, fmt.Errorf("GetUserByName:  %w", ErrFailedExecuteQuery)
	}

	r.logger.Info("User found", "user_id", user.ID)
	return &user, nil
}

// GetUserByID получение пользователя по id
func (r *UserRepo) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User

	r.logger.Info("Executing query", "query", queryGetUserByID, "user_id", id)
	err := r.db.QueryRow(ctx, queryGetUserByID, id).Scan(&user.ID, &user.UserName, &user.Password, &user.Balance, &user.UpdateBalance, &user.Referrer, &user.CreatedAt)
	if err != nil {
		r.logger.Error("Failed to execute query to get user by id", "error", err, "user_id", id)
		return nil, fmt.Errorf("GetUserByID: %w", ErrFailedExecuteQuery)
	}

	r.logger.Info("User found", "user_id", user.ID)
	return &user, nil
}

// GetUserLeaderboard получение списка топ пользователей
func (r *UserRepo) GetUserLeaderboard(ctx context.Context) ([]dto.UserLeaderDTO, error) {
	userLeaders := make([]dto.UserLeaderDTO, 0, 10)

	r.logger.Info("Executing query", "query", queryGetLeaderboard)
	rows, err := r.db.Query(ctx, queryGetLeaderboard)
	if err != nil {
		r.logger.Error("Failed to execute query to get leaderboard", "error", err)
		return userLeaders, fmt.Errorf("GetUserLeaderboard: %w", ErrFailedExecuteQuery)
	}
	defer rows.Close()

	for rows.Next() {
		var user dto.UserLeaderDTO
		if err = rows.Scan(&user.ID, &user.UserName, &user.Balance); err != nil {
			r.logger.Error("Failed to parse row", "error", err)
			return userLeaders, fmt.Errorf("GetUserLeaderboard: failed to parse rows: %w", err)
		}
		userLeaders = append(userLeaders, user)
	}
	if err = rows.Err(); err != nil {
		r.logger.Error("Error during rows iteration", "error", err)
		return userLeaders, fmt.Errorf("GetUserLeaderboard: error during rows iteration: %w", err)
	}

	r.logger.Info("Leaderboard gotten")
	return userLeaders, nil
}

// SetReferrer добавление реферера пользователю
func (r *UserRepo) SetReferrer(ctx context.Context, tx pgx.Tx, userID, referrer int) error {
	r.logger.Info("Executing query", "query", queryUpdateReferrer, "user_id", userID, "referrer", referrer)

	_, err := tx.Exec(ctx, queryUpdateReferrer, referrer, userID)
	if err != nil {
		r.logger.Error("Failed to set referrer", "error", err, "user_id", userID, "referrer", referrer)
		return fmt.Errorf("SetReferrer:  %w", ErrFailedExecuteQuery)
	}

	r.logger.Info("Referrer set", "user_id", userID, "referrer", referrer)
	return nil
}

// AddPoint добавление поинтов за выполенение задания
func (r *UserRepo) AddPoint(ctx context.Context, tx pgx.Tx, userID int, points int) error {
	r.logger.Info("Executing query", "query", queryUpdatePoints, "user_id", userID)

	_, err := tx.Exec(ctx, queryUpdatePoints, points, userID)
	if err != nil {
		r.logger.Error("Failed to add points", "error", err, "user_id", userID)
		return fmt.Errorf("AddPoint:  %w", ErrFailedExecuteQuery)
	}
	r.logger.Info("Points added and updated_at updated", "user_id", userID)
	return nil
}

// GetTask получение данных о задании
func (r *UserRepo) GetTask(ctx context.Context, taskID int) (*models.Task, error) {
	var storedTask models.Task

	r.logger.Info("Executing query", "query", queryGetTask, "taskID", taskID)
	err := r.db.QueryRow(ctx, queryGetTask, taskID).Scan(&storedTask.ID, &storedTask.Description, &storedTask.Reward)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Info("Task not found", "taskID", taskID)
			return nil, fmt.Errorf("GetTask:  %w", ErrTaskNotFound)
		}

		r.logger.Error("Failed to execute query to get task", "error", err, "taskID", taskID)
		return nil, fmt.Errorf("GetTask:  %w", ErrFailedExecuteQuery)
	}

	r.logger.Info("Task found", "task_id", storedTask.ID, "description", storedTask.Description)
	return &storedTask, nil
}

// IsCompletedTask проверка выполнения задания
func (r *UserRepo) IsCompletedTask(ctx context.Context, tx pgx.Tx, userID, taskID int) (bool, error) {
	r.logger.Info("Executing query", "query", queryIsCompletedTask, "userID", userID, "taskID", taskID)

	var exist bool
	err := tx.QueryRow(ctx, queryIsCompletedTask, userID, taskID).Scan(&exist)
	if err != nil && err != pgx.ErrNoRows {
		r.logger.Error("Failed to execute query to check completed task", "error", err, "userID", userID, "taskID", taskID)
		return false, fmt.Errorf("IsCompletedTask:  %w", ErrFailedExecuteQuery)
	}

	r.logger.Info("Task already completed", "user_id", userID, "taskID", taskID)
	return exist, nil
}

// AddCompletedTask добавление выполненной задачи
func (r *UserRepo) AddCompletedTask(ctx context.Context, tx pgx.Tx, userID, taskID int) error {
	r.logger.Info("Executing query", "query", queryCompletedTask, "taskID", taskID)

	_, err := tx.Exec(ctx, queryCompletedTask, userID, taskID)
	if err != nil {
		r.logger.Error("Failed to execute query to add completed task", "error", err, "userID", userID, "taskID", taskID)
		return fmt.Errorf("AddCompletedTask:  %w", ErrFailedExecuteQuery)
	}

	r.logger.Info("Completed task added", "user_id", userID, "taskID", taskID)
	return nil
}
