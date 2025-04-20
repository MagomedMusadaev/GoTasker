package useCase

import (
	"context"
	"log/slog"
	"time"
)

type TaskBackRepository interface {
	DeleteExpiredTasks(ctx context.Context) (int64, error)
}
type BackgroundJob struct {
	taskRepository TaskBackRepository
}

func NewBackgroundJob(taskRepository TaskBackRepository) *BackgroundJob {
	return &BackgroundJob{
		taskRepository: taskRepository,
	}
}

func (b *BackgroundJob) StartTaskCleanup(taskRepo TaskBackRepository, tickerTime int) {
	const op = "internal.useCase.background_jons.StartTaskCleanup"

	ticker := time.NewTicker(time.Hour * time.Duration(tickerTime))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			deletedCount, err := taskRepo.DeleteExpiredTasks(context.Background())
			if err != nil {
				slog.Error(op, "ошибка удаления просроченных задач", slog.String("err", err.Error()))
			} else {
				slog.Info("удаление просроченных задач прошло успешно", slog.Int64("deleted", deletedCount))
			}
		}
	}
}
