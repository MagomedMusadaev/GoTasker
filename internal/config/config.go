package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config содержит все конфигурационные параметры приложения
type Config struct {
	DB     DBConfig     // Настройки базы данных
	Server ServerConfig // Настройки сервера
	Log    LogConfig    // Настройки логирования
	Redis  RedisConfig  // Настройки Redis
	Env    string       // Текущее окружение (development, production, test)
}

// DBConfig содержит параметры подключения к базе данных
type DBConfig struct {
	Host     string // Хост базы данных
	Port     int    // Порт базы данных
	User     string // Имя пользователя
	Password string // Пароль
	DBName   string // Название базы данных
	SSLMode  string // Режим SSL-подключения
}

// ServerConfig содержит настройки HTTP-сервера
type ServerConfig struct {
	Port            string        // Порт для HTTP-сервера
	JWTSecret       string        // Ключ для JWT токена
	RefreshDuration time.Duration // ttl refresh токена
	AccessDuration  time.Duration // ttl access токена
	TaskCleanupDays int           // Время для фоновой джобы очиски задач
}

// RedisConfig содержит настройки подключения к Redis
type RedisConfig struct {
	Host     string        // Хост Redis
	Port     int           // Порт Redis
	Password string        // Пароль Redis
	DB       int           // Номер базы данных Redis
	TTL      time.Duration // Время жизни кэша
}

// LogConfig содержит настройки логирования
type LogConfig struct {
	Level       string // Уровень логирования (DEBUG, INFO, WARN, ERROR)
	FilePath    string // Путь к файлу логов
	Environment string // Окружение для формата логов
}

// GetLogDir возвращает директорию для логов
func (c *LogConfig) GetLogDir() string {
	return filepath.Dir(c.FilePath)
}

// NewConfig создает новый экземпляр конфигурации с учетом переменных окружения
func NewConfig() *Config {
	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("не удалось получить рабочую директорию: %s", err))
	}

	rootDir := workDir
	for {
		if _, err = os.Stat(filepath.Join(rootDir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(rootDir)
		if parent == rootDir {
			panic("не удалось найти корневую директорию проекта (go.mod не найден)")
		}
		rootDir = parent
	}

	if err = godotenv.Load(filepath.Join(rootDir, ".env")); err != nil {
		panic(fmt.Sprintf(".env файл не найден: %s", err))
	}

	config := &Config{
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", "password123"),
			DB:       getEnvAsInt("REDIS_DB", 0),
			TTL:      time.Duration(getEnvAsInt("REDIS_CACHE_TTL", 15)) * time.Minute,
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", ":8080"),
			JWTSecret:       getEnv("JWT_SECRET", "secret"),
			AccessDuration:  time.Duration(getEnvAsInt("ACCESS_DURATION", 15)),  // minutes
			RefreshDuration: time.Duration(getEnvAsInt("REFRESH_DURATION", 30)), // days
			TaskCleanupDays: getEnvAsInt("TASK_CLEANUP_DAYS", 7),
		},
		Log: LogConfig{
			Level:       getEnv("LOG_LEVEL", "INFO"),
			FilePath:    filepath.Join(rootDir, getEnv("LOG_FILE", "logs/app.log")),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Env: getEnv("ENVIRONMENT", "development"),
	}

	if err = config.Validate(); err != nil {
		panic(fmt.Sprintf("невалидная конфигурация: %v", err))
	}

	return config
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	// Проверка настроек БД
	if c.DB.Host == "" {
		return fmt.Errorf("хост базы данных не может быть пустым")
	}
	if c.DB.Port <= 0 || c.DB.Port > 65535 {
		return fmt.Errorf("недопустимый порт базы данных")
	}
	if c.DB.User == "" {
		return fmt.Errorf("имя пользователя базы данных не может быть пустым")
	}

	// Проверка настроек сервера
	if c.Server.Port == "" {
		return fmt.Errorf("порт сервера не может быть пустым")
	}

	// Проверка настроек логирования
	validLogLevels := map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true}
	if !validLogLevels[strings.ToUpper(c.Log.Level)] {
		return fmt.Errorf("недопустимый уровень логирования: %s", c.Log.Level)
	}

	// Проверка окружения
	validEnvs := map[string]bool{"development": true, "production": true, "test": true}
	if !validEnvs[c.Env] {
		return fmt.Errorf("недопустимое окружение: %s", c.Env)
	}

	return nil
}

// GetConnectionString формирует строку подключения к базе данных
func (c *DBConfig) GetConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Вспомогательные функции для работы с переменными окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := fmt.Sscanf(value, "%d"); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}
