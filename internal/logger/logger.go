package logger

import (
	"GoTasker/internal/config"
	"context"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"os"
	"strings"
)

// MultiHandler обработчик для множественного вывода логов
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler создает новый мультиплексор для обработки логов
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// Handle реализует интерфейс slog.Handler
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// WithAttrs реализует интерфейс slog.Handler
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewMultiHandler(handlers...)
}

// WithGroup реализует интерфейс slog.Handler
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewMultiHandler(handlers...)
}

// Enabled реализует интерфейс slog.Handler
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// NewLogger создает новый логгер с настроенной ротацией и уровнем логирования
func NewLogger(cfg *config.LogConfig) *slog.Logger {
	// Настройка ротации логов
	logWriter := &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    10, // 10 MB
		MaxBackups: 3,  // до 3 файлов
		MaxAge:     7,  // 7 дней
		Compress:   true,
	}

	// Определение уровня логирования
	var level slog.Level
	switch strings.ToUpper(cfg.Level) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Создаем мультиплексор для вывода в файл и консоль
	writers := []slog.Handler{}

	// В режиме разработки используем текстовый формат
	if cfg.Environment == "development" {
		// Обработчик для файла
		fileHandler := slog.NewTextHandler(logWriter, &slog.HandlerOptions{Level: level})
		// Обработчик для консоли
		consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		writers = append(writers, fileHandler, consoleHandler)
	} else {
		// Обработчик для файла
		fileHandler := slog.NewJSONHandler(logWriter, &slog.HandlerOptions{Level: level})
		// Обработчик для консоли
		consoleHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		writers = append(writers, fileHandler, consoleHandler)
	}

	// Создаем мультиплексор
	handler := NewMultiHandler(writers...)

	return slog.New(handler)
}

// SetupGlobalLogger устанавливает глобальный логгер
func SetupGlobalLogger(cfg *config.LogConfig) error {
	logger := NewLogger(cfg)
	slog.SetDefault(logger)

	// Проверка возможности записи в файл
	if err := os.MkdirAll(cfg.GetLogDir(), 0755); err != nil {
		return fmt.Errorf("не удалось создать журнал для логов: %w", err)
	}

	return nil
}
