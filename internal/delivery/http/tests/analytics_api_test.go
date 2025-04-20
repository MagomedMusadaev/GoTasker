package tests

import (
	"GoTasker/internal/domain"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalyticsAPI_GetTaskCountByStatus(t *testing.T) {
	s := setupTestServer(t)

	t.Run("успешное получение статистики по статусам", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/status-count", nil)

		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]int
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Проверка, что в ответе есть хотя бы один статус с количеством задач
		assert.NotEmpty(t, response)
		for status, count := range response {
			assert.Greater(t, count, 0, "Для статуса %s количество задач должно быть больше 0", status)
		}
	})
}

func TestAnalyticsAPI_GetAverageExecutionTime(t *testing.T) {
	s := setupTestServer(t)

	t.Run("успешное получение среднего времени выполнения", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/average-execution-time", nil)

		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			AverageTime string `json:"average_time"`
		}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AverageTime)
	})
}

func TestAnalyticsAPI_GetReportPeriod(t *testing.T) {
	s := setupTestServer(t)

	t.Run("успешное получение отчета за период", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/report", nil)

		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.ReportPeriod
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Проверяем, что поля отчета не пустые и имеют смысл
		assert.GreaterOrEqual(t, response.CompletedTasks, 0)
		assert.GreaterOrEqual(t, response.OverdueTasks, 0)
	})
}
