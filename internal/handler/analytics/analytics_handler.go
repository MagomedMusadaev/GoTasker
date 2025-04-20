package analytics

import (
	"GoTasker/internal/domain"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TaskAnalyticsUseCase interface {
	GetAnalytics(ctx context.Context) (*domain.AnalyticsTasksResponse, error)
}

type TaskAnalyticsHandler struct {
	taskAnalyticsUseCase TaskAnalyticsUseCase
}

func NewAnalyticsHandler(taskAnalyticsUseCase TaskAnalyticsUseCase) *TaskAnalyticsHandler {
	return &TaskAnalyticsHandler{
		taskAnalyticsUseCase: taskAnalyticsUseCase,
	}
}

// @Summary Получение аналитики
// @Description Возвращает аналитические данные по задачам
// @Tags Аналитика
// @Produce json
// @Success 200 {object} domain.AnalyticsTasksResponse
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /analytics [get]
// @Security bearerAuth
func (h *TaskAnalyticsHandler) GetAnalytics(c *gin.Context) {
	const op = "internal.handler.analytics_handler.GetAnalytics"

	ctx := c.Request.Context()

	analytics, err := h.taskAnalyticsUseCase.GetAnalytics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить аналитику",
		})

		return
	}

	c.JSON(http.StatusOK, analytics)
}
