package tasks

import (
	"context"
	"testing"
	"time"

	"GoTasker/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskPostgresRepo struct {
	mock.Mock
}

func (m *MockTaskPostgresRepo) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskPostgresRepo) Update(ctx context.Context, updates map[string]interface{}) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

func (m *MockTaskPostgresRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskPostgresRepo) GetAll(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskPostgresRepo) ImportTasks(ctx context.Context, tasks []*domain.Task) (int, error) {
	args := m.Called(ctx, tasks)
	return args.Int(0), args.Error(1)
}

func TestTaskPostgresRepo_Create(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := new(MockTaskPostgresRepo)

	// Задача для теста
	task := &domain.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      domain.StatusPending,
		Priority:    domain.PriorityLow,
		DueDate:     time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Настроим ожидание: метод Create должен быть вызван и вернуть nil (успех)
	mockRepo.On("Create", mock.Anything, task).Return(nil)

	// Выполняем метод Create
	err := mockRepo.Create(context.Background(), task)

	// Проверяем, что ошибок нет
	assert.NoError(t, err)

	// Проверяем, что метод был вызван с ожидаемыми аргументами
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_Create_Error(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	// Задача для теста
	task := &domain.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      domain.StatusPending,
		Priority:    domain.PriorityLow,
		DueDate:     time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Настроим мок, чтобы метод Create вернул ошибку
	mockRepo.On("Create", mock.Anything, task).Return(assert.AnError)

	// Выполняем метод Create
	err := mockRepo.Create(context.Background(), task)

	// Проверяем, что ошибка произошла
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)

	// Проверяем, что метод был вызван с ожидаемыми аргументами
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_Create_NoPriority(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	// Задача без приоритета (по умолчанию будет `low`)
	task := &domain.Task{
		Title:       "Test Task Without Priority",
		Description: "Test Description Without Priority",
		Status:      domain.StatusPending,
		DueDate:     time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Настроим мок, чтобы метод Create возвращал nil (успех)
	mockRepo.On("Create", mock.Anything, task).Return(nil)

	// Выполняем метод Create
	err := mockRepo.Create(context.Background(), task)

	// Проверяем, что ошибок нет
	assert.NoError(t, err)

	// Проверяем, что метод был вызван с ожидаемыми аргументами
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_Update(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	updates := map[string]interface{}{
		"title": "Updated Task Title",
	}

	// Ожидаем, что метод Update будет вызван и вернет nil (успех)
	mockRepo.On("Update", mock.Anything, updates).Return(nil)

	err := mockRepo.Update(context.Background(), updates)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_Update_Error(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	updates := map[string]interface{}{
		"title": "Updated Task Title",
	}

	// Настроим мок, чтобы метод Update вернул ошибку
	mockRepo.On("Update", mock.Anything, updates).Return(assert.AnError)

	err := mockRepo.Update(context.Background(), updates)

	// Проверяем, что ошибка произошла
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_Delete(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	// Ожидаем, что метод Delete будет вызван с id 1 и вернет nil (успех)
	mockRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

	err := mockRepo.Delete(context.Background(), 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_Delete_Error(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	// Настроим мок, чтобы метод Delete вернул ошибку
	mockRepo.On("Delete", mock.Anything, int64(1)).Return(assert.AnError)

	err := mockRepo.Delete(context.Background(), 1)

	// Проверяем, что ошибка произошла
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_GetAll(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	// Создаем фильтр и задачи
	filter := &domain.TaskFilter{}
	tasks := []*domain.Task{
		{Title: "Task 1", Status: domain.StatusPending},
		{Title: "Task 2", Status: domain.StatusInProgress},
	}

	// Ожидаем, что метод GetAll вернет список задач
	mockRepo.On("GetAll", mock.Anything, filter).Return(tasks, nil)

	result, err := mockRepo.GetAll(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, tasks, result)
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_GetAll_Error(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	// Создаем фильтр
	filter := &domain.TaskFilter{}

	// Настроим мок, чтобы метод GetAll вернул пустой срез и ошибку
	mockRepo.On("GetAll", mock.Anything, filter).Return([]*domain.Task{}, assert.AnError)

	// Выполняем метод GetAll
	result, err := mockRepo.GetAll(context.Background(), filter)

	// Проверяем, что ошибка произошла
	assert.Error(t, err)
	// Проверяем, что результат — это пустой срез
	assert.Empty(t, result)

	// Проверяем, что метод был вызван с ожидаемыми аргументами
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_GetAll_FilterByStatus(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	// Создаем фильтр для задач со статусом "pending"
	filter := &domain.TaskFilter{Status: string(domain.StatusPending)}
	tasks := []*domain.Task{
		{Title: "Task 1", Status: domain.StatusPending},
	}

	// Ожидаем, что метод GetAll вернет только задачи со статусом "pending"
	mockRepo.On("GetAll", mock.Anything, filter).Return(tasks, nil)

	result, err := mockRepo.GetAll(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, tasks, result)
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_ImportTasks(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	// Задачи для импорта
	tasks := []*domain.Task{
		{Title: "Task 1", Status: domain.StatusPending},
		{Title: "Task 2", Status: domain.StatusInProgress},
	}

	// Ожидаем, что метод ImportTasks вернет количество импортированных задач
	mockRepo.On("ImportTasks", mock.Anything, tasks).Return(2, nil)

	count, err := mockRepo.ImportTasks(context.Background(), tasks)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	mockRepo.AssertExpectations(t)
}

func TestTaskPostgresRepo_ImportTasks_Error(t *testing.T) {
	mockRepo := new(MockTaskPostgresRepo)

	// Задачи для импорта
	tasks := []*domain.Task{
		{Title: "Task 1", Status: domain.StatusPending},
		{Title: "Task 2", Status: domain.StatusInProgress},
	}

	// Настроим мок, чтобы метод ImportTasks вернул ошибку
	mockRepo.On("ImportTasks", mock.Anything, tasks).Return(0, assert.AnError)

	count, err := mockRepo.ImportTasks(context.Background(), tasks)

	// Проверяем, что ошибка произошла
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	mockRepo.AssertExpectations(t)
}
