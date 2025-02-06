package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-management/internal/config"
	"user-management/internal/database"
	"user-management/internal/delivery"
	"user-management/internal/middleware"
	"user-management/internal/pkg/logger"
	_ "user-management/internal/pkg/validation"
	"user-management/internal/repository"
	"user-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type App struct {
	dbConn         *pgx.Conn
	config         *config.Config
	logger         *slog.Logger
	apiServer      *http.Server
	userService    service.UserService
	userHandler    delivery.UserHandler
	tokenService   service.TokenService
	authMiddleware *middleware.AuthMiddleware
}

func New() (*App, error) {
	app := &App{}

	// Загружаем конфигурацию
	config := config.MustLoad()

	// Инициализируем логгер
	logger := logger.InitLogger(slog.LevelDebug)

	// Подключаемся к базе данных
	dbConn, err := database.NewDBConnection(&config.DatabaseConfig)
	if err != nil {
		logger.Error("Failed to connect to the database", "error", err)
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	// Инициализация репозитория
	userRepo := repository.NewUserRepository(dbConn, logger)
	tokenRepo := repository.NewTokenRepo(dbConn, logger)

	// Инициализация сервисного слоя
	userService := service.NewUserService(userRepo, logger)
	tokenService := service.NewTokenService(tokenRepo, config.ApiServerConfig.AuthSecretKey, logger)

	// Инициализация обработчиков
	userHandler := delivery.NewUserHandler(userService, tokenService, config, logger)

	// Инициализация middleware
	authMiddleware := middleware.NewAuthMiddleware(tokenService, logger)

	// Собираем приложение
	app.config = config
	app.logger = logger
	app.dbConn = dbConn
	app.userService = userService
	app.userHandler = userHandler
	app.tokenService = tokenService
	app.authMiddleware = authMiddleware

	// Настраиваем API
	apiRouter := gin.Default()
	app.configureApiRoutes(apiRouter)

	// Формируем адрес для сервера из конфигурации
	apiAddress := fmt.Sprintf("%s:%s", app.config.ApiServerConfig.Host, app.config.ApiServerConfig.Port)

	// Инициализация сервера
	app.apiServer = &http.Server{
		Addr:    apiAddress,
		Handler: apiRouter,
	}

	logger.Info("Application initialized successfully")
	return app, nil
}

func (app *App) Run() error {
	// Канал для сигналов о завершении
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP-сервера
	go func() {
		app.logger.Info("API server started successfully:", "address", app.apiServer.Addr)
		if err := app.apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Error("Failed to start the server", "error", err)
		}
	}()

	// Ожидаем сигнал завершения
	sig := <-stop
	app.logger.Info("Received shutdown signal", "signal", sig)

	// Пытаемся корректно завершить работу сервера.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Закрываем HTTP сервер
	if err := app.apiServer.Shutdown(ctx); err != nil {
		app.logger.Error("HTTP server shutdown failed", "error", err)
		return err
	}

	// Закрываем соединение с базой данных
	app.Close()

	app.logger.Info("Application stopped gracefully")
	return nil
}

// Close закрывает соединение с базой данных
func (app *App) Close() {
	if app.dbConn != nil {
		if err := app.dbConn.Close(context.Background()); err != nil {
			app.logger.Error("Failed to close database connection", "error", err)
		} else {
			app.logger.Info("Database connection closed successfully")
		}
	}
}
