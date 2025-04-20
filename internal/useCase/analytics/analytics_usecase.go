package analytics

import (
	"GoTasker/internal/domain"
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
)

type TaskAnalyticsRepository interface {
	GetTaskCountByStatus(ctx context.Context) (map[string]int, error)
	GetAverageExecutionTime(ctx context.Context) (string, error)
	GetReportPeriod(ctx context.Context) (*domain.ReportPeriod, error)
}

type RedisRepoAnalytics interface {
	GetAnalytics(ctx context.Context) (*domain.AnalyticsTasksResponse, error)
	SetAnalytics(ctx context.Context, analytics *domain.AnalyticsTasksResponse) error
}

type TaskAnalyticsUseCase struct {
	taskRepository TaskAnalyticsRepository
	redisRepo      RedisRepoAnalytics
}

func NewAnalyticsUseCase(taskRepository TaskAnalyticsRepository, redisRepo RedisRepoAnalytics) *TaskAnalyticsUseCase {
	return &TaskAnalyticsUseCase{
		taskRepository: taskRepository,
		redisRepo:      redisRepo,
	}
}

func (uc *TaskAnalyticsUseCase) GetAnalytics(ctx context.Context) (*domain.AnalyticsTasksResponse, error) {
	const op = "internal.useCase.analytics_useCase.GetAnalytics"

	// Пробуем получить данные из кэша
	cachedAnalytics, err := uc.redisRepo.GetAnalytics(ctx)
	if err == nil && cachedAnalytics != nil {
		return cachedAnalytics, nil
	}

	// 1. Получаем количество задач по статусам
	statusCounts, err := uc.taskRepository.GetTaskCountByStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить количество задач по статусам: %w", err)
	}

	// 2. Получаем среднее время выполнения задач
	avgExecutionTime, err := uc.taskRepository.GetAverageExecutionTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить среднее время выполнения задач: %w", err)
	}

	finalAvgExecutionTime, err := formatExecutionTime(avgExecutionTime)
	if err != nil {
		slog.Error(op, "ошибка форматирования времени", slog.String("error", err.Error()))
		return nil, err
	}

	// 3. Получаем отчет по задачам за период
	report, err := uc.taskRepository.GetReportPeriod(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить отчет по задачам: %w", err)
	}

	analyticsResponse := &domain.AnalyticsTasksResponse{
		StatusCounts:         statusCounts,
		AverageExecutionTime: finalAvgExecutionTime,
		ReportLastPeriod:     report,
	}

	if err = uc.redisRepo.SetAnalytics(ctx, analyticsResponse); err != nil {
		slog.Error(op, "ошибка сохранения данных в кэш", slog.String("err", err.Error()))
	}

	return analyticsResponse, nil
}

func formatExecutionTime(timeStr string) (string, error) {
	re := regexp.MustCompile(`(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?`)
	matches := re.FindStringSubmatch(timeStr)

	if len(matches) == 0 {
		return "", fmt.Errorf("не удалось распарсить строку времени: %s", timeStr)
	}

	hours := 0
	if matches[1] != "" {
		hours, _ = strconv.Atoi(matches[1])
	}

	minutes := 0
	if matches[2] != "" {
		minutes, _ = strconv.Atoi(matches[2])
	}

	seconds := 0
	if matches[3] != "" {
		seconds, _ = strconv.Atoi(matches[3])
	}

	return fmt.Sprintf("%d часов %d минут %d секунд", hours, minutes, seconds), nil
}
