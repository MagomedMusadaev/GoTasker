package domain

import "time"

type Status string
type Priority string

const (
	StatusPending    Status = "pending"     // Задача в ожидании, еще не начата.
	StatusInProgress Status = "in_progress" // Задача в процессе выполнения.
	StatusDone       Status = "done"        // Задача выполнена.

	PriorityLow    Priority = "low"    // Задача с низким приоритетом.
	PriorityMedium Priority = "medium" // Задача с средним приоритетом.
	PriorityHigh   Priority = "high"   // Задача с высоким приоритетом.
)

// Task представляет задачу с различными атрибутами.
type Task struct {
	ID          int64     `json:"id,omitempty" db:"id"`                   // Уникальный идентификатор задачи в базе данных (auto increment).
	Title       string    `json:"title,omitempty" db:"title"`             // Название задачи.
	Description string    `json:"description,omitempty" db:"description"` // Описание задачи (опционально).
	Status      Status    `json:"status,omitempty" db:"status"`           // Статус задачи (значения: pending, in_progress, done).
	Priority    Priority  `json:"priority,omitempty" db:"priority"`       // Приоритет задачи (значения: low, medium, high).
	DueDate     time.Time `json:"due_date" db:"due_date"`                 // Дата завершения задачи.
	CreatedAt   time.Time `json:"created_at" db:"created_at"`             // Дата создания задачи в базе данных.
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`             // Дата последнего обновления задачи в базе данных.
}

// CreateTaskRequest сугубо для swagger
type CreateTaskRequest struct {
	Title       string `json:"title" example:"task 1"`
	Description string `json:"description" example:"info"`
	Priority    string `json:"priority" example:"low"`
	Status      string `json:"status" example:"pending"`
	DueDate     string `json:"due_date" example:"2025-05-03T00:00:00Z"`
}

// TaskFilter структура для фильтрации задач
type TaskFilter struct {
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
	DueDate  string `json:"due_date,omitempty"`
	Title    string `json:"title,omitempty"`
}

func NewTaskFilter(status string, priority string, dueDate string, title string) *TaskFilter {
	return &TaskFilter{
		Status:   status,
		Priority: priority,
		DueDate:  dueDate,
		Title:    title,
	}
}

// AnalyticsTasksResponse структура для сбора аналитики задач
type AnalyticsTasksResponse struct {
	StatusCounts         map[string]int `json:"status_counts"`
	AverageExecutionTime string         `json:"average_execution_time"`
	ReportLastPeriod     *ReportPeriod  `json:"report_last_period"`
}

// ReportPeriod структура для хранения количества завершённых и просроченных задач за указанный период
type ReportPeriod struct {
	CompletedTasks int `json:"completed_tasks"`
	OverdueTasks   int `json:"overdue_tasks"`
}
