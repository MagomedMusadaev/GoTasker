package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	_ "GoTasker/docs"
	"GoTasker/internal/config"
	"GoTasker/internal/delivery/http"
	"GoTasker/internal/logger"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	// Handlers
	analyticsHandler "GoTasker/internal/handler/analytics"
	authHandler "GoTasker/internal/handler/auth"
	tasksHandler "GoTasker/internal/handler/tasks"

	// Repositories
	tasksRepo "GoTasker/internal/repository/postgres/tasks"
	usersRepo "GoTasker/internal/repository/postgres/users"
	"GoTasker/internal/repository/redis"

	// UseCases
	analyticsUC "GoTasker/internal/useCase/analytics"
	authUC "GoTasker/internal/useCase/auth"
	tasksUC "GoTasker/internal/useCase/tasks"

	"GoTasker/internal/useCase"
)

// @title GoTasker API
// @version 1.0
// @description API для управления задачами
// @host localhost:8085
// @BasePath /
// @securityDefinitions.apikey bearerAuth
// @in header
// @name Authorization
func main() {
	const op = "cmd.app.main"

	// Загрузка конфигурации
	cfg := config.NewConfig()

	// Настройка логирования
	if err := logger.SetupGlobalLogger(&cfg.Log); err != nil {
		fmt.Printf("Не удалось настроить logger: %v\n", err)
		os.Exit(1)
	}

	slog.Info("Запуск приложения", "environment", cfg.Env)

	// Установка соединения с базой данных Psql
	db, err := sql.Open("postgres", cfg.DB.GetConnectionString())
	if err != nil {
		slog.Error(op, "Ошибка подключения к базе данных:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Проверка соединения с базой данных
	if err = db.Ping(); err != nil {
		slog.Error(op, "Ошибка при проверке соединения с базой данных:", err)
		os.Exit(1)
	}

	// Репозитории
	taskRepo := tasksRepo.NewTaskPostgresRepo(db)
	userRepo := usersRepo.NewUserPostgresRepo(db)
	analyticsRedis := redis.NewAnalyticsRedisRepo(cfg)

	// UseCases
	taskUC := tasksUC.NewTaskUseCase(taskRepo)
	authUseCase := authUC.NewAuthUseCase(userRepo, cfg)
	analyticUC := analyticsUC.NewAnalyticsUseCase(taskRepo, analyticsRedis)
	backgroundJob := useCase.NewBackgroundJob(taskRepo)

	// Handlers
	taskHand := tasksHandler.NewTaskHandler(taskUC)
	authHand := authHandler.NewUserAuthHandler(authUseCase)
	analyticHand := analyticsHandler.NewAnalyticsHandler(analyticUC)

	// Маршруты
	r := gin.Default()
	http.SetupRoutes(r, taskHand, analyticHand, authHand)

	// Запуск фоновых задач
	go backgroundJob.StartTaskCleanup(taskRepo, cfg.Server.TaskCleanupDays)

	// Старт сервера
	slog.Warn(fmt.Sprintf("Сервер запущен и прослушивает порт %s\n", cfg.Server.Port))
	if err = r.Run(cfg.Server.Port); err != nil {
		log.Fatalf("ошибка запуска сервера: %v", err)
	}
}
