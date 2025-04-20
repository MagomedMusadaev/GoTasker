package tasks

import (
	"GoTasker/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type TaskPostgresRepo struct {
	db *sql.DB
}

func NewTaskPostgresRepo(db *sql.DB) *TaskPostgresRepo {
	return &TaskPostgresRepo{
		db: db,
	}
}

func (r *TaskPostgresRepo) Create(ctx context.Context, task *domain.Task) error {
	const op = "internal.repository.postgres.task_repo.Create"

	query := `
		INSERT INTO tasks (title, description, status, priority, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`

	if err := r.db.QueryRowContext(
		ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
		task.CreatedAt,
		task.UpdatedAt).Scan(&task.ID); err != nil {
		slog.Error(op, "не удалось сохранить задачу",
			slog.String("title", task.Title),
			slog.String("status", string(task.Status)),
			slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *TaskPostgresRepo) Update(ctx context.Context, updates map[string]interface{}) error {
	const op = "internal.repository.postgres.task_repo.Update"

	taskID := updates["id"]
	var exists bool

	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)", taskID).Scan(&exists)
	if err != nil {
		slog.Error(op, "ошибка при проверке существования задачи", slog.String("err", err.Error()))
		return fmt.Errorf("не удалось проверить существование задачи: %w", err)
	}
	if !exists {
		return fmt.Errorf("задача с id %v не найдена", taskID)
	}

	var setParts []string
	var args []interface{}
	i := 1

	for k, v := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}

	args = append(args, taskID)
	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d", strings.Join(setParts, ", "), i)

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		slog.Error(op, "ошибка обновления задачи", slog.String("err", err.Error()))
		return fmt.Errorf("не удалось обновить задачу: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.Error(op, "не удалось получить количество затронутых строк", slog.String("err", err.Error()))
		return fmt.Errorf("не удалось проверить количество затронутых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача с id %v не найдена", taskID)
	}

	return nil
}

func (r *TaskPostgresRepo) Delete(ctx context.Context, id int64) error {
	const op = "internal.repository.postgres.task_repo.Delete"

	query := `DELETE FROM tasks WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		slog.Error(op, "не удалось удалить задачу", slog.String("err", err.Error()))
		return err
	}

	// Проверка, была ли задача удалена
	affect, _ := res.RowsAffected()
	if affect == 0 {
		err = fmt.Errorf("задача с id %d не найдена для удаления", id)
		slog.Error(op, "задача для удаления не найдена", slog.String("err", err.Error()))
		return err
	}

	return nil
}

func (r *TaskPostgresRepo) GetAll(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	const op = "internal.repository.postgres.task_repo.GetAll"

	query := `
		SELECT id, title, description, status, priority, due_date, created_at, updated_at 
		FROM tasks 
	`

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, filter.Priority)
		argIdx++
	}

	if filter.DueDate != "" {
		conditions = append(conditions, fmt.Sprintf("due_date = $%d", argIdx))
		args = append(args, filter.DueDate)
		argIdx++
	}

	if filter.Title != "" {
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Title+"%")
		argIdx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error(op, "не удалось получить задачи", slog.String("err", err.Error()))
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var task domain.Task
		if err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			slog.Error(op, "не удалось извлечь данные задачи", slog.String("err", err.Error()))
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	slog.Info("получено задач", slog.Int("count", len(tasks)))

	if err = rows.Err(); err != nil {
		slog.Error(op, "ошибка при переборе строк", slog.String("err", err.Error()))
		return nil, err
	}

	return tasks, nil
}

func (r *TaskPostgresRepo) DeleteExpiredTasks(ctx context.Context) (int64, error) {
	const op = "internal.repository.postgres.task_repo.DeleteExpiredTasks"

	query := `DELETE FROM tasks WHERE due_date < NOW() - INTERVAL '7 days'`

	res, err := r.db.ExecContext(ctx, query)
	if err != nil {
		slog.Error(op, "не удалось удалить просроченные задачи", slog.String("err", err.Error()))
		return 0, fmt.Errorf("не удалось удалить просроченные задачи: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected, nil
}

func (r *TaskPostgresRepo) ImportTasks(ctx context.Context, tasks []*domain.Task) (int, error) {
	const op = "internal.repository.postgres.task_repo.ImportTasks"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("не удалось начать транзакцию: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var values []string
	var args []interface{}

	for i, task := range tasks {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))
		args = append(args, task.Title, task.Description, task.Status, task.Priority, task.DueDate, task.CreatedAt, task.UpdatedAt)
	}

	query := fmt.Sprintf(`
        INSERT INTO tasks (title, description, status, priority, due_date, created_at, updated_at)
        VALUES %s
    `, strings.Join(values, ", "))

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		slog.Error(op, "ошибка импорта задач", slog.String("err", err.Error()))
		return 0, fmt.Errorf("не удалось импортировать задачи: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("не удалось зафиксировать транзакцию: %w", err)
	}

	return len(values), nil
}

func (r *TaskPostgresRepo) GetTaskCountByStatus(ctx context.Context) (map[string]int, error) {
	const op = "internal.repository.postgres.task_repo.GetAnalytics"

	query := `SELECT status, COUNT(*) FROM tasks GROUP BY status`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error(op, "ошибка выполнения запроса", slog.String("err", err.Error()))
		return nil, err
	}
	defer rows.Close()

	statusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err = rows.Scan(&status, &count); err != nil {
			slog.Error(op, "ошибка при сканировании строки", slog.String("err", err.Error()))
			return nil, err
		}
		statusCounts[status] = count
	}

	slog.Info("успешное получение количества задач по статусам", slog.Int("total_statuses", len(statusCounts)))
	return statusCounts, nil
}

func (r *TaskPostgresRepo) GetAverageExecutionTime(ctx context.Context) (string, error) {
	const op = "internal.repository.postgres.task_repo.GetAverageExecutionTime"

	query := `SELECT AVG(EXTRACT(EPOCH FROM (due_date - created_at))) FROM tasks WHERE status = 'done'`

	var avgSeconds sql.NullFloat64 // используем sql.NullFloat64 для обработки NULL значений
	err := r.db.QueryRowContext(ctx, query).Scan(&avgSeconds)
	if err != nil {
		slog.Error(
			op,
			"ошибка выполнения запроса для среднего времени выполнения задач",
			slog.String("err", err.Error()),
		)
		return "", err
	}

	// Проверяем, что значение не NULL
	if !avgSeconds.Valid {
		avgSeconds.Float64 = 0 // если NULL, то ставим 0
	}

	avgDuration := time.Duration(avgSeconds.Float64 * float64(time.Second))
	slog.Info(
		"успешное получение среднего времени выполнения задач",
		slog.String("avg_execution_time", avgDuration.String()),
	)

	return avgDuration.String(), nil
}

func (r *TaskPostgresRepo) GetReportPeriod(ctx context.Context) (*domain.ReportPeriod, error) {
	const op = "internal.repository.postgres.task_repo.GetReportPeriod"

	query := `
		SELECT COUNT(*) 
		FROM tasks 
		WHERE updated_at >= NOW() - INTERVAL '7 days' AND status = 'done'
	`

	var completed, overdue int
	err := r.db.QueryRowContext(ctx, query).Scan(&completed)
	if err != nil {
		slog.Error(
			op,
			"ошибка получения количества завершённых задач за последние 7 дней",
			slog.String("err", err.Error()),
		)

		return nil, err
	}

	query = `
		SELECT COUNT(*) 
		FROM tasks 
		WHERE due_date < NOW() - INTERVAL '7 days' AND status != 'done'
	`

	err = r.db.QueryRowContext(ctx, query).Scan(&overdue)
	if err != nil {
		slog.Error(
			op,
			"ошибка получения количества просроченных задач за последние 7 дней",
			slog.String("err", err.Error()),
		)

		return nil, err
	}

	slog.Info(op, "успешно получен отчёт за последние 7 дней",
		slog.Int("completed_tasks", completed),
		slog.Int("overdue_tasks", overdue))

	return &domain.ReportPeriod{
		CompletedTasks: completed,
		OverdueTasks:   overdue,
	}, nil
}
