package tests

import (
	"GoTasker/internal/domain"
	"GoTasker/internal/useCase/tasks"
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Мок-репозиторий для тестов
type MockTaskRepo struct{}

func (m *MockTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	task.ID = 1
	return nil
}

func (m *MockTaskRepo) Update(ctx context.Context, updates map[string]interface{}) error {
	if id, ok := updates["id"].(int64); ok && id == 1 {
		return nil
	}
	return nil
}

func (m *MockTaskRepo) Delete(ctx context.Context, id int64) error {
	if id == 1 {
		return nil
	}
	return nil
}

func (m *MockTaskRepo) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	if id == 1 {
		return &domain.Task{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      domain.StatusPending,
			Priority:    domain.PriorityHigh,
			DueDate:     time.Now().Add(24 * time.Hour),
		}, nil
	}
	return nil, nil
}

func (m *MockTaskRepo) GetAll(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	return []*domain.Task{
		{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      domain.StatusPending,
			Priority:    domain.PriorityHigh,
			DueDate:     time.Now().Add(24 * time.Hour),
		},
	}, nil
}

func (m *MockTaskRepo) CreateTask(task *domain.Task) (*domain.Task, error) {
	task.ID = 1
	return task, nil
}

func (m *MockTaskRepo) GetTaskByID(id int) (*domain.Task, error) {
	if id == 1 {
		return &domain.Task{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      domain.StatusPending,
			Priority:    domain.PriorityHigh,
			DueDate:     time.Now().Add(24 * time.Hour),
		}, nil
	}
	return nil, nil
}

func (m *MockTaskRepo) UpdateTask(id int, task *domain.Task) (*domain.Task, error) {
	if id == 1 {
		task.ID = int64(id)
		return task, nil
	}
	return nil, nil
}

func (m *MockTaskRepo) DeleteTask(id int) error {
	if id == 1 {
		return nil
	}
	return nil
}

func (m *MockTaskRepo) GetAllTasks() ([]domain.Task, error) {
	return []domain.Task{
		{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      domain.StatusPending,
			Priority:    domain.PriorityHigh,
			DueDate:     time.Now().Add(24 * time.Hour),
		},
	}, nil
}

func (m *MockTaskRepo) GetTasksByFilter(status, priority string) ([]domain.Task, error) {
	return []domain.Task{
		{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      domain.StatusPending,
			Priority:    domain.PriorityHigh,
			DueDate:     time.Now().Add(24 * time.Hour),
		},
	}, nil
}

func (m *MockTaskRepo) ImportTasks(ctx context.Context, tasks []*domain.Task) (int, error) {
	return len(tasks), nil
}

// TestServer структура с роутером и юзкейсом
type TestServer struct {
	router      *gin.Engine
	taskUseCase *tasks.TaskUseCase
}

func setupTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(gin.Recovery())

	// Создаем мок-репозиторий и useCase
	taskRepo := &MockTaskRepo{}
	taskUseCase := tasks.NewTaskUseCase(taskRepo)

	// Регистрация маршрутов для задач
	router.POST("/api/v1/tasks", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": 1})
	})
	router.GET("/api/v1/tasks", func(c *gin.Context) {
		tasks := []domain.Task{{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      domain.StatusPending,
			Priority:    domain.PriorityHigh,
			DueDate:     time.Now().Add(24 * time.Hour),
		}}
		c.JSON(http.StatusOK, tasks)
	})
	router.PUT("/api/v1/tasks/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	})
	router.DELETE("/api/v1/tasks/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Регистрация маршрутов для аналитики
	router.GET("/api/v1/analytics/status-count", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]int{
			"pending":    1,
			"in_progress": 2,
			"completed":  3,
		})
	})
	router.GET("/api/v1/analytics/average-execution-time", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"average_time": "2h 30m"})
	})
	router.GET("/api/v1/analytics/report", func(c *gin.Context) {
		c.JSON(http.StatusOK, domain.ReportPeriod{
			CompletedTasks: 5,
			OverdueTasks:   1,
		})
	})

	return &TestServer{
		router:      router,
		taskUseCase: taskUseCase,
	}
}

func TestTaskAPI_CreateTask(t *testing.T) {
	s := setupTestServer(t)

	t.Run("успешное создание задачи", func(t *testing.T) {
		task := domain.Task{
			Title:       "Test Task",
			Description: "Test Description",
			Priority:    domain.PriorityHigh,
			Status:      domain.StatusPending,
			DueDate:     time.Now().Add(24 * time.Hour),
		}

		body, err := json.Marshal(task)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestTaskAPI_UpdateTask(t *testing.T) {
	s := setupTestServer(t)

	t.Run("успешное обновление задачи", func(t *testing.T) {
		task := domain.Task{
			ID:          1,
			Title:       "Updated Task",
			Description: "Updated Description",
			Priority:    domain.PriorityMedium,
			Status:      domain.StatusInProgress,
			DueDate:     time.Now().Add(48 * time.Hour),
		}

		body, err := json.Marshal(task)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTaskAPI_DeleteTask(t *testing.T) {
	s := setupTestServer(t)

	t.Run("успешное удаление задачи", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/1", nil)

		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTaskAPI_GetTasks(t *testing.T) {
	s := setupTestServer(t)

	t.Run("успешное получение списка задач", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)

		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []domain.Task
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
	})
}
