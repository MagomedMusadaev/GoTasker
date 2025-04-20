package analytics

import (
	"GoTasker/internal/domain"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockTaskAnalyticsUseCase struct {
	mock.Mock
}

func (m *MockTaskAnalyticsUseCase) GetAnalytics(ctx context.Context) (*domain.AnalyticsTasksResponse, error) {
	args := m.Called(ctx)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	return val.(*domain.AnalyticsTasksResponse), args.Error(1)
}

func TestTaskAnalyticsUseCase_GetAnalytics(t *testing.T) {
	mockUseCase := new(MockTaskAnalyticsUseCase)

	t.Run("успешное получение аналитики", func(t *testing.T) {
		expectedResponse := &domain.AnalyticsTasksResponse{
			StatusCounts: map[string]int{
				string(domain.StatusPending):    5,
				string(domain.StatusInProgress): 3,
				string(domain.StatusDone):       2,
			},
			AverageExecutionTime: "2h 30m",
			ReportLastPeriod: &domain.ReportPeriod{
				CompletedTasks: 5,
				OverdueTasks:   4,
			},
		}

		mockUseCase.On("GetAnalytics", mock.Anything).Return(expectedResponse, nil)

		response, err := mockUseCase.GetAnalytics(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("ошибка при получении аналитики", func(t *testing.T) {
		mockUseCase.ExpectedCalls = nil

		mockUseCase.On("GetAnalytics", mock.Anything).Return(nil, assert.AnError)

		response, err := mockUseCase.GetAnalytics(context.Background())

		assert.Error(t, err)
		assert.Nil(t, response)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("отсутствие данных", func(t *testing.T) {
		mockUseCase.ExpectedCalls = nil

		expectedResponse := &domain.AnalyticsTasksResponse{
			StatusCounts:         make(map[string]int),
			AverageExecutionTime: "",
			ReportLastPeriod:     nil,
		}

		mockUseCase.On("GetAnalytics", mock.Anything).Return(expectedResponse, nil)

		response, err := mockUseCase.GetAnalytics(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("некорректный формат времени", func(t *testing.T) {
		mockUseCase.ExpectedCalls = nil

		expectedResponse := &domain.AnalyticsTasksResponse{
			StatusCounts: map[string]int{
				string(domain.StatusPending):    5,
				string(domain.StatusInProgress): 3,
				string(domain.StatusDone):       2,
			},
			AverageExecutionTime: "invalid_time_format", // некорректный формат
			ReportLastPeriod: &domain.ReportPeriod{
				CompletedTasks: 5,
				OverdueTasks:   4,
			},
		}

		mockUseCase.On("GetAnalytics", mock.Anything).Return(expectedResponse, nil)

		response, err := mockUseCase.GetAnalytics(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		assert.Equal(t, "invalid_time_format", response.AverageExecutionTime)

		mockUseCase.AssertExpectations(t)
	})

	t.Run("пустой отчет за период", func(t *testing.T) {
		mockUseCase.ExpectedCalls = nil

		expectedResponse := &domain.AnalyticsTasksResponse{
			StatusCounts: map[string]int{
				string(domain.StatusPending):    0,
				string(domain.StatusInProgress): 0,
				string(domain.StatusDone):       0,
			},
			AverageExecutionTime: "0h 0m 0s",
			ReportLastPeriod: &domain.ReportPeriod{
				CompletedTasks: 0,
				OverdueTasks:   0,
			},
		}

		mockUseCase.On("GetAnalytics", mock.Anything).Return(expectedResponse, nil)

		response, err := mockUseCase.GetAnalytics(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		mockUseCase.AssertExpectations(t)
	})
}
