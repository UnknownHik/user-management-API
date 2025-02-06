package logger

import (
	"log/slog"
	"os"
)

// InitLogger создает и возвращает новый логгер
func InitLogger(level slog.Level) *slog.Logger {
	// Кастомизация обработчика
	options := &slog.HandlerOptions{
		Level: level,
	}
	// Создает обработчик, который выводит логи в консоль.
	handler := slog.NewTextHandler(os.Stdout, options)

	return slog.New(handler)
}
