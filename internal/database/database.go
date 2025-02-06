package database

import (
	"context"
	"fmt"
	"log"

	"user-management/internal/config"

	"github.com/jackc/pgx/v5"
)

// NewDBConnection устанавливает подключение к базе данных с использованием строки подключения
func NewDBConnection(dbcfg *config.Database) (*pgx.Conn, error) {
	// Получаем строку подключения
	connStr := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		dbcfg.Driver,
		dbcfg.User,
		dbcfg.Password,
		dbcfg.Host,
		dbcfg.DBPort,
		dbcfg.Db,
		dbcfg.SSLMode,
	)

	// Устанавливаем соединение с базой данных
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
		return nil, err
	}

	return conn, nil
}
