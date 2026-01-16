package config

import (
	"fmt"

	"github.com/paincake00/geocore/internal/env"
)

// Config хранит настройки приложения.
type Config struct {
	httpPort    string
	databaseURL string
	redisAddr   string
	webhookURL  string
	apiKey      string
	statsWindow int
}

// Load загружает конфигурацию из переменных окружения.
func Load() *Config {
	return &Config{
		httpPort:    env.GetString("HTTP_PORT", "8080"),
		databaseURL: getDatabaseURL(),
		redisAddr:   getRedisAddr(),
		webhookURL:  env.GetString("WEBHOOK_URL", "http://localhost:9090"),
		apiKey:      env.GetString("API_KEY", ""), // пустое значение по умолчанию
		statsWindow: env.GetInt("STATS_TIME_WINDOW_MINUTES", 30),
	}
}

// Геттеры для доступа к приватным полям конфигурации
func (c *Config) HTTPPort() string    { return c.httpPort }
func (c *Config) DatabaseURL() string { return c.databaseURL }
func (c *Config) RedisAddr() string   { return c.redisAddr }
func (c *Config) WebhookURL() string  { return c.webhookURL }
func (c *Config) APIKey() string      { return c.apiKey }
func (c *Config) StatsWindow() int    { return c.statsWindow }

// getDatabaseURL формирует строку подключения к PostgreSQL.
func getDatabaseURL() string {
	// Если DATABASE_URL задан явно (например, в docker-compose), используем его.
	// Иначе собираем из компонентов.
	if url := env.GetString("DATABASE_URL", ""); url != "" {
		return url
	}

	driver := env.GetString("DB_DRIVER", "postgres")
	user := env.GetString("POSTGRES_USER", "user")
	password := env.GetString("POSTGRES_PASSWORD", "password")
	host := env.GetString("POSTGRES_HOST", "localhost")
	port := env.GetString("POSTGRES_PORT", "5432")
	dbName := env.GetString("POSTGRES_DB", "geocore")

	return fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", driver, user, password, host, port, dbName)
}

// getRedisAddr формирует адрес подключения к Redis.
func getRedisAddr() string {
	if addr := env.GetString("REDIS_ADDR", ""); addr != "" {
		return addr
	}
	host := env.GetString("REDIS_HOST", "localhost")
	port := env.GetString("REDIS_PORT", "6379")
	return fmt.Sprintf("%s:%s", host, port)
}
