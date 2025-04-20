package tasks

import (
	"GoTasker/internal/domain"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type mockTaskRepo struct {
	mock.Mock
}

func (m *mockTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockTaskRepo) Update(ctx context.Context, updates map[string]interface{}) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

func (m *mockTaskRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTaskRepo) GetAll(ctx context.Context, f *domain.TaskFilter) ([]*domain.Task, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *mockTaskRepo) ImportTasks(ctx context.Context, tasks []*domain.Task) (int, error) {
	args := m.Called(ctx, tasks)
	return args.Int(0), args.Error(1)
}

func TestTaskUseCase_Create(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockTaskRepo)
	uc := NewTaskUseCase(mockRepo)

	t.Run("успешное создание задачи", func(t *testing.T) {
		task := &domain.Task{
			Title:    "Test Task",
			Priority: "high",
			Status:   "pending",
			DueDate:  time.Now().Add(24 * time.Hour),
		}
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Task")).Return(nil)

		err := uc.Create(ctx, task)
		assert.NoError(t, err)
		mockRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*domain.Task"))
	})

	t.Run("ошибка валидации - пустой title", func(t *testing.T) {
		task := &domain.Task{
			Title:    "",
			Priority: "high",
			Status:   "pending",
			DueDate:  time.Now().Add(24 * time.Hour),
		}

		err := uc.Create(ctx, task)
		assert.ErrorContains(t, err, "название задачи не может быть пустым")
	})

	t.Run("ошибка валидации - невалидный приоритет", func(t *testing.T) {
		task := &domain.Task{
			Title:    "Test Task",
			Priority: "crazy", // невалидный
			Status:   "pending",
			DueDate:  time.Now().Add(24 * time.Hour),
		}

		err := uc.Create(ctx, task)
		assert.ErrorContains(t, err, "некорректный приоритет задачи")
	})
}

func TestTaskUseCase_Update(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockTaskRepo)
	uc := NewTaskUseCase(mockRepo)

	t.Run("успешное обновление задачи", func(t *testing.T) {
		task := &domain.Task{
			ID:       1,
			Title:    "Updated Task",
			Priority: "medium",
			Status:   "in_progress",
			DueDate:  time.Now().Add(48 * time.Hour),
		}
		mockRepo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.Update(ctx, task)
		assert.NoError(t, err)
		mockRepo.AssertCalled(t, "Update", ctx, mock.Anything)
	})

	t.Run("ошибка валидации - пустой ID", func(t *testing.T) {
		task := &domain.Task{
			Title:    "Updated Task",
			Priority: "medium",
			Status:   "in_progress",
			DueDate:  time.Now().Add(48 * time.Hour),
		}

		err := uc.Update(ctx, task)
		assert.ErrorContains(t, err, "id задачи не может быть нулевым")
	})

	t.Run("ошибка валидации - невалидный статус", func(t *testing.T) {
		task := &domain.Task{
			ID:       1,
			Title:    "Updated Task",
			Priority: "medium",
			Status:   "invalid_status", // невалидный статус
			DueDate:  time.Now().Add(48 * time.Hour),
		}

		err := uc.Update(ctx, task)
		assert.ErrorContains(t, err, "невалидный статус задачи")
	})
}

func TestTaskUseCase_Delete(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockTaskRepo)
	uc := NewTaskUseCase(mockRepo)

	t.Run("успешное удаление задачи", func(t *testing.T) {
		mockRepo.On("Delete", ctx, int64(1)).Return(nil)

		err := uc.Delete(ctx, 1)
		assert.NoError(t, err)
		mockRepo.AssertCalled(t, "Delete", ctx, int64(1))
	})

	t.Run("ошибка валидации - нулевой ID", func(t *testing.T) {
		err := uc.Delete(ctx, 0)
		assert.ErrorContains(t, err, "id задачи не может быть нулевым")
	})
}

func TestTaskUseCase_GetAll(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockTaskRepo)
	uc := NewTaskUseCase(mockRepo)

	t.Run("успешное получение всех задач", func(t *testing.T) {
		mockRepo.On("GetAll", ctx, mock.Anything).Return([]*domain.Task{
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
		}, nil)

		tasks, err := uc.GetAll(ctx, &domain.TaskFilter{})
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		mockRepo.AssertCalled(t, "GetAll", ctx, mock.Anything)
	})
}

func TestTaskUseCase_Import(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockTaskRepo)
	uc := NewTaskUseCase(mockRepo)

	t.Run("успешный импорт задач", func(t *testing.T) {
		tasks := []*domain.Task{
			{Title: "Task 1", Priority: "medium", Status: "pending", DueDate: time.Now().Add(24 * time.Hour)},
			{Title: "Task 2", Priority: "high", Status: "done", DueDate: time.Now().Add(48 * time.Hour)},
		}

		mockRepo.On("ImportTasks", ctx, mock.MatchedBy(func(tasksArg []*domain.Task) bool {
			if len(tasksArg) != len(tasks) {
				return false
			}
			for _, task := range tasksArg {
				if task.Title == "" || task.Priority == "" || task.Status == "" {
					return false
				}
			}
			return true
		})).Return(2, nil)

		inserted, invalidTasks, err := uc.Import(ctx, tasks)
		assert.NoError(t, err)
		assert.Equal(t, 2, inserted)
		assert.Len(t, invalidTasks, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при импорте задач", func(t *testing.T) {
		tasks := []*domain.Task{
			{Title: "Task 1", Priority: "medium", Status: "pending", DueDate: time.Now().Add(24 * time.Hour)},
		}
		mockRepo.On("ImportTasks", ctx, tasks).Return(0, fmt.Errorf("import error"))

		inserted, invalidTasks, err := uc.Import(ctx, tasks)
		assert.Error(t, err)
		assert.Equal(t, 0, inserted)
		assert.Len(t, invalidTasks, 0)
		mockRepo.AssertCalled(t, "ImportTasks", ctx, tasks)
	})

	t.Run("ошибка валидации при импорте", func(t *testing.T) {
		tasks := []*domain.Task{
			{Title: "Task 1", Priority: "crazy", Status: "pending", DueDate: time.Now().Add(24 * time.Hour)}, // невалидный приоритет
		}

		inserted, invalidTasks, err := uc.Import(ctx, tasks)
		assert.Error(t, err)
		assert.Len(t, invalidTasks, 1)
		assert.Equal(t, "Задача 1: некорректный приоритет задачи: crazy", invalidTasks[0])
		assert.Equal(t, 0, inserted)
	})
}
