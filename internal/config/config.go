package config

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config представляет конфигурацию приложения
type Config struct {
	ApiServerConfig ApiServer
	DatabaseConfig  Database
}

// ApiServer представляет конфигурацию сервера API
type ApiServer struct {
	Host          string        `env:"API_SERVER_HOST" env-default:"localhost"`
	Port          string        `env:"API_SERVER_PORT" env-default:"8080"`
	AuthSecretKey string        `env:"API_SERVER_AUTH_SECRET_KEY" env-required:"true"`
	Timeout       time.Duration `env:"API_SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout   time.Duration `env:"API_SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

// Database представляет конфигурацию подключения к базе данных
type Database struct {
	Driver   string `env:"DB_DRIVER" env-default:"postgres"`   // Драйвер подключения к БД
	Db       string `env:"DB_NAME" env-default:"postgres"`     // Имя БД
	User     string `env:"DB_USER" env-default:"postgres"`     // Имя пользователя БД
	Password string `env:"DB_PASSWORD" env-default:"postgres"` // Пароль пользователя БД
	Host     string `env:"DB_HOST" env-default:"db"`           // Хост БД
	DBPort   string `env:"DB_PORT" env-default:"5432"`         // Порт БД
	SSLMode  string `env:"DB_SSLMODE" env-default:"disable"`   // Режим SSL для подключения к БД

}

var (
	cfg  *Config
	once sync.Once
)

// MustLoad загружает конфигурацию
func MustLoad() *Config {
	once.Do(func() {
		// Инициализируем конфигурацию
		cfg = &Config{}

		// Загружаем конфигурацию для сервера из переменных окружения
		if err := cleanenv.ReadConfig(".env", &cfg.ApiServerConfig); err != nil {
			log.Fatalf("Failed to load API server configuration from env: %s", err)
		}

		// Загружаем конфигурацию для базы данных из переменных окружения
		if err := cleanenv.ReadConfig(".env", &cfg.DatabaseConfig); err != nil {
			log.Fatalf("Failed to load database configuration from env: %s", err)
		}

		log.Println("Config loaded successfully...")
	})

	return cfg
}
