package tasks

import (
	"GoTasker/internal/domain"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type TaskPostgresRepo interface {
	Create(ctx context.Context, task *domain.Task) error
	Update(ctx context.Context, updates map[string]interface{}) error
	Delete(ctx context.Context, id int64) error
	GetAll(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error)
	ImportTasks(ctx context.Context, tasks []*domain.Task) (int, error)
}

type TaskUseCase struct {
	taskRepository TaskPostgresRepo
}

func NewTaskUseCase(taskRepository TaskPostgresRepo) *TaskUseCase {
	return &TaskUseCase{
		taskRepository: taskRepository,
	}
}

func (uc *TaskUseCase) Create(ctx context.Context, task *domain.Task) error {
	const op = "internal.useCase.task_useCase.Create"

	if err := validateTask(task); err != nil {
		slog.Error(op, "ошибка валидации", slog.String("err", err.Error()))
		return err
	}

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	return uc.taskRepository.Create(ctx, task)
}

func (uc *TaskUseCase) Update(ctx context.Context, updatedTask *domain.Task) error {
	const op = "internal.useCase.task_useCase.Update"

	updates := make(map[string]interface{})

	if updatedTask.Title != "" {
		updates["title"] = updatedTask.Title
	}
	if updatedTask.Description != "" {
		updates["description"] = updatedTask.Description
	}
	if updatedTask.Status != "" {
		if !isValidStatus(string(updatedTask.Status)) {
			err := fmt.Errorf("невалидный статус задачи: %s", updatedTask.Status)
			slog.Error(op, "ошибка валидации", slog.String("err", err.Error()))
			return err
		}
		updates["status"] = updatedTask.Status
	}
	if updatedTask.Priority != "" {
		if !isValidPriority(string(updatedTask.Priority)) {
			err := fmt.Errorf("невалидный приоритет задачи: %s", updatedTask.Priority)
			slog.Error(op, "ошибка валидации", slog.String("err", err.Error()))
			return err
		}
		updates["priority"] = updatedTask.Priority
	}
	if !updatedTask.DueDate.IsZero() {
		updates["due_date"] = updatedTask.DueDate
	}

	if len(updates) == 0 {
		return fmt.Errorf("нет данных для обновления")
	}

	if updatedTask.ID == 0 {
		err := fmt.Errorf("id задачи не может быть нулевым")
		slog.Error(op, "ошибка валидации", slog.String("err", err.Error()))
		return err
	}
	updates["id"] = updatedTask.ID
	updates["updated_at"] = time.Now()

	return uc.taskRepository.Update(ctx, updates)
}

func (uc *TaskUseCase) Delete(ctx context.Context, id int64) error {
	const op = "internal.useCase.task_useCase.Delete"

	if id == 0 {
		err := fmt.Errorf("id задачи не может быть нулевым")
		slog.Error(op, "ошибка валидации", slog.String("err", err.Error()))
		return err
	}

	return uc.taskRepository.Delete(ctx, id)
}

func (uc *TaskUseCase) GetAll(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	const op = "internal.useCase.task_useCase.GetAll"

	return uc.taskRepository.GetAll(ctx, filter)
}

func (uc *TaskUseCase) Import(ctx context.Context, tasks []*domain.Task) (int, []string, error) {
	const op = "internal.useCase.task_useCase.Import"

	var validTasks []*domain.Task
	var invalidTasks []string
	var wg sync.WaitGroup

	taskChan := make(chan struct {
		task  *domain.Task
		err   error
		index int
	}, len(tasks))

	for i, task := range tasks {
		wg.Add(1)
		go func(i int, task *domain.Task) {
			defer wg.Done()
			// Проверка валидности задачи
			err := validateTask(task)
			taskChan <- struct {
				task  *domain.Task
				err   error
				index int
			}{task, err, i}
		}(i, task)
	}

	go func() {
		wg.Wait()
		close(taskChan)
	}()

	for result := range taskChan {
		if result.err != nil {
			invalidTasks = append(invalidTasks, fmt.Sprintf("Задача %d: %s", result.index+1, result.err.Error()))
		} else {
			result.task.CreatedAt = time.Now()
			result.task.UpdatedAt = time.Now()
			validTasks = append(validTasks, result.task)
		}
	}

	if len(validTasks) == 0 {
		return 0, invalidTasks, fmt.Errorf("все задачи невалидны")
	}

	inserted, err := uc.taskRepository.ImportTasks(ctx, validTasks)
	if err != nil {
		return 0, invalidTasks, err
	}

	return inserted, invalidTasks, nil
}

func validateTask(task *domain.Task) error {
	if task.Title == "" {
		return fmt.Errorf("название задачи не может быть пустым")
	}
	if task.Priority == "" {
		return fmt.Errorf("приоритет задачи не может быть пустым")
	}
	if task.DueDate.IsZero() {
		return fmt.Errorf("не указана дата завершения задачи")
	}
	if task.Status != "" && task.Status != "pending" && task.Status != "in_progress" && task.Status != "done" {
		return fmt.Errorf("невалидный статус задачи")
	}
	if !isValidStatus(string(task.Status)) {
		return fmt.Errorf("некорректный статус задачи: %s", task.Status)
	}

	if !isValidPriority(string(task.Priority)) {
		return fmt.Errorf("некорректный приоритет задачи: %s", task.Priority)
	}

	return nil
}

func isValidStatus(status string) bool {
	validStatuses := []string{"pending", "in_progress", "done"}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func isValidPriority(priority string) bool {
	validPriorities := []string{"low", "medium", "high"}
	for _, p := range validPriorities {
		if priority == p {
			return true
		}
	}
	return false
}
