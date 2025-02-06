package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"user-management/internal/dto"
	"user-management/internal/models"
	"user-management/internal/repository"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Ошибки сервисного слоя
var (
	ErrInvalidReferrer = errors.New("invalid referrer")
	ErrSetReferrer     = errors.New("user has referrer")
	ErrIsCompletedTask = errors.New("task already completed")
)

type UserService interface {
	Register(ctx context.Context, user *dto.UserRegLogDTO) (int, error)
	Login(ctx context.Context, user *dto.UserRegLogDTO) (*dto.UserLoginDTO, error)
	UserStatus(ctx context.Context, userID int) (*dto.UserStatusDTO, error)
	UserLeaderboard(ctx context.Context) ([]dto.UserLeaderDTO, error)
	AddReferrer(ctx context.Context, userID int, referrer *dto.ReferrerDTO) error
	TaskComplete(ctx context.Context, userID int, task *dto.TaskDTO) error
}

type DefaultUserService struct {
	repo   repository.UserRepository
	logger *slog.Logger
}

func NewUserService(repo repository.UserRepository, logger *slog.Logger) *DefaultUserService {
	return &DefaultUserService{repo: repo, logger: logger}
}

// Register регистрирует нового пользователя
func (s *DefaultUserService) Register(ctx context.Context, userDTO *dto.UserRegLogDTO) (userID int, err error) {
	s.logger.Info("Starting user registration", "username", userDTO.UserName)

	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", "error", err)
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	s.logger.Info("Transaction started")

	defer s.handleTransaction(ctx, tx, &err)

	existingUser, err := s.repo.GetUserByNameWithTx(ctx, tx, userDTO.UserName)

	if existingUser != nil {
		s.logger.Warn("User already exists", "username", userDTO.UserName)
		tx.Rollback(ctx)
		return 0, fmt.Errorf("user with username %s already exists", userDTO.UserName)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return 0, err
	}

	user := &models.User{
		UserName: userDTO.UserName,
		Password: string(hashedPassword),
	}

	userID, err = s.repo.CreateUserWithTx(ctx, tx, user)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return 0, fmt.Errorf("error creating user: %w", err)
	}

	s.logger.Info("User created successfully", "user_id", userID)
	return userID, nil
}

// handleTransaction управляет коммитом или откатом транзакции
func (s *DefaultUserService) handleTransaction(ctx context.Context, tx pgx.Tx, err *error) {
	if p := recover(); p != nil {
		tx.Rollback(ctx)
		panic(p)
	} else if *err != nil {
		tx.Rollback(ctx)
	} else {
		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			*err = fmt.Errorf("failed to commit transaction: %w", commitErr)
		} else {
			s.logger.Info("Transaction committed")
		}
	}
}

// Login производит вход пользователя
func (s *DefaultUserService) Login(ctx context.Context, userDTO *dto.UserRegLogDTO) (*dto.UserLoginDTO, error) {
	s.logger.Info("User login attempt", "username", userDTO.UserName)

	storedUser, err := s.repo.GetUserByName(ctx, userDTO.UserName)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err)
		return nil, fmt.Errorf("Login: error getting user: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(userDTO.Password)); err != nil {
		s.logger.Warn("Incorrect password", "username", userDTO.UserName)
		return nil, fmt.Errorf("incorrect password: %w", err)
	}

	s.logger.Info("User logged in successfully", "user_id", storedUser.ID)
	return &dto.UserLoginDTO{ID: storedUser.ID, UserName: storedUser.UserName}, nil
}

// UserStatus предоставляет информацию о пользователе
func (s *DefaultUserService) UserStatus(ctx context.Context, userID int) (*dto.UserStatusDTO, error) {
	s.logger.Info("Fetching user status", "userID", userID)

	storedUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err)
		return nil, fmt.Errorf("UserStatus: error getting user: %w", err)
	}

	return &dto.UserStatusDTO{
		ID:             storedUser.ID,
		UserName:       storedUser.UserName,
		Balance:        storedUser.Balance,
		UpdatedBalance: storedUser.UpdateBalance,
		Referrer:       storedUser.Referrer,
		CreatedAt:      storedUser.CreatedAt,
	}, nil
}

// UserLeaderboard предоставляет топ пользователей с большим балансом
func (s *DefaultUserService) UserLeaderboard(ctx context.Context) ([]dto.UserLeaderDTO, error) {
	s.logger.Info("Fetching user leaderboard")

	leaderboard, err := s.repo.GetUserLeaderboard(ctx)
	if err != nil {
		s.logger.Error("Failed to get leaderboard", "error", err)
		return nil, fmt.Errorf("UserLeaderboard: error getting leaderboard: %w", err)
	}

	return leaderboard, nil
}

// AddReferrer добавляет реферальный код (ID пользователя)
func (s *DefaultUserService) AddReferrer(ctx context.Context, userID int, ref *dto.ReferrerDTO) (err error) {
	s.logger.Info("Starting to add referrer")

	referrerID := ref.Referrer
	if referrerID == userID {
		s.logger.Warn("User ID matches referrer", "userID", userID, "referrer", referrerID)
		return ErrInvalidReferrer
	}

	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	s.logger.Info("Transaction started")

	defer s.handleTransaction(ctx, tx, &err)

	storedUser, err := s.repo.GetUserByIDWithTx(ctx, tx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err)
		return fmt.Errorf("error getting user: %w", err)
	}

	if storedUser.Referrer != nil {
		s.logger.Warn("User already has referrer", "userID", userID, "referrer", storedUser.Referrer)
		tx.Rollback(ctx)
		return ErrSetReferrer
	}

	storedReferrer, err := s.repo.GetUserByIDWithTx(ctx, tx, referrerID)
	if err != nil {
		s.logger.Error("Failed to search referrer", "error", err)
		return fmt.Errorf("error searching referrer: %w", err)
	}

	if storedReferrer.Referrer != nil && *storedReferrer.Referrer == userID {
		s.logger.Warn("Detected circular referral", "userID", userID, "referrer", storedReferrer.Referrer)
		tx.Rollback(ctx)
		return ErrInvalidReferrer
	}

	err = s.repo.SetReferrer(ctx, tx, userID, referrerID)
	if err != nil {
		s.logger.Error("Failed to set referrer", "error", err)
		return fmt.Errorf("error setting referrer: %w", err)
	}

	pointsForRef := 80
	err = s.repo.AddPoint(ctx, tx, referrerID, pointsForRef)
	if err != nil {
		s.logger.Error("Failed to add points for referral", "error", err)
		return fmt.Errorf("error adding points: %w", err)
	}

	s.logger.Info("Referrer added successful")
	return nil
}

// TaskComplete определяет выполнение задания
func (s *DefaultUserService) TaskComplete(ctx context.Context, userID int, task *dto.TaskDTO) error {
	s.logger.Info("Starting to complete task")

	storedTask, err := s.repo.GetTask(ctx, task.ID)
	if err != nil {
		s.logger.Error("Failed to get task", "error", err)
		return fmt.Errorf("error getting task: %w", err)
	}

	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	s.logger.Info("Transaction started")

	defer s.handleTransaction(ctx, tx, &err)

	isCompleted, err := s.repo.IsCompletedTask(ctx, tx, userID, storedTask.ID)
	if err != nil {
		s.logger.Error("Failed to check completed task", "error", err)
		return fmt.Errorf("error checking completed task: %w", err)
	}
	if isCompleted {
		s.logger.Warn("Task already completed by user", "user_id", userID, "task_id", storedTask.ID)
		return ErrIsCompletedTask
	}

	err = s.repo.AddPoint(ctx, tx, userID, storedTask.Reward)
	if err != nil {
		s.logger.Error("Failed to add points for task", "error", err)
		return fmt.Errorf("error adding points: %w", err)
	}

	err = s.repo.AddCompletedTask(ctx, tx, userID, storedTask.ID)
	if err != nil {
		s.logger.Error("Failed to add completed task", "error", err)
		return fmt.Errorf("error adding completed task: %w", err)
	}

	s.logger.Info("Task completed successful")
	return nil
}
