package tasks

import (
	"GoTasker/internal/domain"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockTaskAnalyticsRepo struct {
	mock.Mock
}

// GetTaskCountByStatus мок-метод для получения количества задач по статусам
func (m *MockTaskAnalyticsRepo) GetTaskCountByStatus(ctx context.Context) (map[string]int, error) {
	args := m.Called(ctx)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	return val.(map[string]int), args.Error(1)
}

// GetAverageExecutionTime мок-метод для получения среднего времени выполнения
func (m *MockTaskAnalyticsRepo) GetAverageExecutionTime(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// GetReportPeriod мок-метод для получения отчета за период
func (m *MockTaskAnalyticsRepo) GetReportPeriod(ctx context.Context) (*domain.ReportPeriod, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ReportPeriod), args.Error(1)
}

func newMockRepo() *MockTaskAnalyticsRepo {
	return new(MockTaskAnalyticsRepo)
}

func TestTaskAnalyticsRepo_GetTaskCountByStatus(t *testing.T) {
	mockRepo := newMockRepo()

	t.Run("успешное получение статистики по статусам", func(t *testing.T) {
		expectedCounts := map[string]int{
			string(domain.StatusPending):    5,
			string(domain.StatusInProgress): 3,
			string(domain.StatusDone):       2,
		}

		mockRepo.On("GetTaskCountByStatus", mock.Anything).Return(expectedCounts, nil)

		counts, err := mockRepo.GetTaskCountByStatus(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedCounts, counts)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении статистики", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil

		mockRepo.On("GetTaskCountByStatus", mock.Anything).Return(nil, assert.AnError)

		// Выполнение метода
		counts, err := mockRepo.GetTaskCountByStatus(context.Background())

		// Проверка результатов
		assert.Error(t, err)
		assert.Nil(t, counts)
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskAnalyticsRepo_GetAverageExecutionTime(t *testing.T) {
	mockRepo := newMockRepo()

	t.Run("успешное получение среднего времени выполнения задач", func(t *testing.T) {
		// Тестовые данные
		expectedTime := "2h 30m"

		// Настройка ожидания
		mockRepo.On("GetAverageExecutionTime", mock.Anything).Return(expectedTime, nil)

		// Выполнение метода
		avgTime, err := mockRepo.GetAverageExecutionTime(context.Background())

		// Проверка результатов
		assert.NoError(t, err)
		assert.Equal(t, expectedTime, avgTime)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении среднего времени выполнения задач", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil

		mockRepo.On("GetAverageExecutionTime", mock.Anything).Return("", assert.AnError)

		avgTime, err := mockRepo.GetAverageExecutionTime(context.Background())

		assert.Error(t, err)
		assert.Empty(t, avgTime)
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskAnalyticsRepo_GetReportPeriod(t *testing.T) {
	mockRepo := newMockRepo()

	t.Run("успешное получение отчета за последние 7 дней", func(t *testing.T) {

		expectedReport := &domain.ReportPeriod{
			CompletedTasks: 5,
			OverdueTasks:   4,
		}

		mockRepo.On("GetReportPeriod", mock.Anything).Return(expectedReport, nil)

		report, err := mockRepo.GetReportPeriod(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedReport, report)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении отчета за последние 7 дней", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil

		mockRepo.On("GetReportPeriod", mock.Anything).Return(nil, assert.AnError)

		report, err := mockRepo.GetReportPeriod(context.Background())

		assert.Error(t, err)
		assert.Nil(t, report)
		mockRepo.AssertExpectations(t)
	})
}
