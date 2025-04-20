package tasks

import (
	"GoTasker/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TaskUseCase interface {
	Create(ctx context.Context, task *domain.Task) error
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id int64) error
	GetAll(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error)
	Import(ctx context.Context, tasks []*domain.Task) (int, []string, error)
}

type TaskHandler struct {
	useCase TaskUseCase
}

func NewTaskHandler(useCase TaskUseCase) *TaskHandler {
	return &TaskHandler{
		useCase: useCase,
	}
}

// @Summary Создание новой задачи
// @Description Создает новую задачу с указанными параметрами
// @Tags Задачи
// @Accept json
// @Produce json
// @Param task body domain.CreateTaskRequest true "Параметры задачи"
// @Success 201 {object} domain.Task
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /tasks [post]
// @Security bearerAuth
func (h *TaskHandler) Create(c *gin.Context) {
	const op = "internal.handler.task_handler.Create"

	var task domain.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		slog.Error(op, "невалидный JSON", slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный JSON"})
		return
	}

	ctx := c.Request.Context()
	if err := h.useCase.Create(ctx, &task); err != nil {
		slog.Error(op, "ошибка создания задачи", slog.String("err", err.Error()))

		if strings.Contains(err.Error(), "название задачи не может быть пустым") ||
			strings.Contains(err.Error(), "приоритет задачи не может быть пустым") ||
			strings.Contains(err.Error(), "не указана дата завершения задачи") ||
			strings.Contains(err.Error(), "невалидный статус задачи") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// @Summary Обновление задачи
// @Description Обновляет существующую задачу по ID
// @Tags Задачи
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Param task body domain.CreateTaskRequest true "Обновленные параметры задачи"
// @Success 200 {object} domain.Task
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /tasks/{id} [put]
// @Security bearerAuth
func (h *TaskHandler) Update(c *gin.Context) {
	const op = "internal.handler.task_handler.Update"

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID задачи не указан"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Error(op, "не удалось преобразовать id", slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный ID задачи"})
		return
	}

	var updatedTask domain.Task
	if err = c.ShouldBindJSON(&updatedTask); err != nil {
		slog.Error(op, "невалидный JSON", slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный JSON"})
		return
	}

	updatedTask.ID = id
	ctx := c.Request.Context()
	if err = h.useCase.Update(ctx, &updatedTask); err != nil {

		customErr := fmt.Sprintf("задача с id %v не найдена", id)
		if strings.Contains(err.Error(), customErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	updatedTask.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, updatedTask)
}

// @Summary Удаление задачи
// @Description Удаляет задачу по ID
// @Tags Задачи
// @Produce json
// @Param id path int true "ID задачи"
// @Success 204 "Задача успешно удалена"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /tasks/{id} [delete]
// @Security bearerAuth
func (h *TaskHandler) Delete(c *gin.Context) {
	const op = "internal.handler.task_handler.Delete"

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID задачи не указан"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Error(op, "не удалось преобразовать id", slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный ID задачи"})
		return
	}

	ctx := c.Request.Context()
	if err = h.useCase.Delete(ctx, id); err != nil {
		slog.Error(op, "ошибка удаления задачи", slog.String("err", err.Error()))

		customErr := fmt.Sprintf("задача с id %d не найдена для удаления", id)
		if strings.Contains(err.Error(), customErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// @Summary Получение списка задач
// @Description Возвращает список всех задач с возможностью фильтрации
// @Tags Задачи
// @Produce json
// @Param status query string false "Фильтр по статусу"
// @Param priority query string false "Фильтр по приоритету"
// @Param due_date query string false "Фильтр по дате завершения"
// @Param title query string false "Фильтр по названию"
// @Success 200 {array} domain.Task
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /tasks [get]
// @Security bearerAuth
func (h *TaskHandler) GetAll(c *gin.Context) {
	const op = "internal.handler.task_handler.GetAll"

	status := c.DefaultQuery("status", "")
	priority := c.DefaultQuery("priority", "")
	dueDate := c.DefaultQuery("due_date", "")
	title := c.DefaultQuery("title", "")

	filter := domain.NewTaskFilter(status, priority, dueDate, title)

	ctx := c.Request.Context()
	tasks, err := h.useCase.GetAll(ctx, filter)
	if err != nil {
		slog.Error(op, "ошибка получения списка задач", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить список задач. Попробуйте позже.",
		})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// @Summary Экспорт задач
// @Description Экспортирует все задачи в JSON файл
// @Tags Задачи
// @Produce json
// @Success 200 {array} domain.Task
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /tasks/export [get]
// @Security bearerAuth
func (h *TaskHandler) Export(c *gin.Context) { // отдаём
	const op = "internal.handler.task_handler.Export"

	ctx := c.Request.Context()
	tasks, err := h.useCase.GetAll(ctx, &domain.TaskFilter{})
	if err != nil {
		slog.Error(op, "ошибка получения списка задач", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить список задач. Попробуйте позже.",
		})
		return
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обработать задачи"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=tasks.json")
	c.Data(http.StatusOK, "application/json", data)
}

// @Summary Импорт задач
// @Description Импортирует задачи из JSON файла
// @Tags Задачи
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "JSON файл с задачами"
// @Success 200 {object} map[string]interface{} "Результат импорта"
// @Failure 400 {object} map[string]string "Ошибка в файле"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /tasks/import [post]
// @Security bearerAuth
func (h *TaskHandler) Import(c *gin.Context) {
	const op = "internal.handler.task_handler.Import"

	ctx := c.Request.Context()

	file, err := c.FormFile("file")
	if err != nil {
		slog.Error("не удалось получить файл", slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "не удалось получить файл"})
		return
	}

	src, err := file.Open()
	if err != nil {
		slog.Error("не удалось открыть файл", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть файл"})
		return
	}
	defer src.Close()

	var tasks []*domain.Task
	if err = json.NewDecoder(src).Decode(&tasks); err != nil {
		slog.Error("ошибка парсинга JSON", slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный JSON формат"})
		return
	}

	inserted, skipped, err := h.useCase.Import(ctx, tasks)
	if err != nil {
		slog.Error("ошибка импорта задач", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         "ошибка при импорте задач",
			"skipped_tasks": skipped,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Импорт успешно завершен",
		"inserted_tasks": inserted,
		"skipped_tasks":  skipped,
	})
}
