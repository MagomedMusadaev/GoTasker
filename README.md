
# GoTasker - Система управления задачами

## Описание проекта
GoTasker - это веб-приложение для управления задачами, разработанное на Go.
Система предоставляет возможности создания, отслеживания и управления задачами с поддержкой аутентификации пользователей и аналитики.

## Основные функции
- Аутентификация пользователей (JWT)
- Управление задачами (CRUD операции)
- Аналитика и отчеты
- Фоновые задачи (автоматическая очистка)
- Кэширование данных (Redis)

## Технологический стек
- Go
- PostgreSQL (основная база данных)
- Redis (кэширование)
- Docker и Docker Compose

## Требования
- Go 1.19 или выше
- PostgreSQL 14+
- Redis 6+
- Docker и Docker Compose (для контейнеризации)

## Установка и запуск

### Локальная разработка
1. Клонируйте репозиторий:
```bash
git clone https://github.com/MagomedMusadaev/GoTasker.git
cd GoTasker
```

2. Создайте файл `.env` на основе примера:
```env
# --- Database settings ---
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=postgres
DB_SSLMODE=disable

# --- Server settings ---
SERVER_PORT=:8085
JWT_SECRET=secret123
ACCESS_DURATION=15
REFRESH_DURATION=30
TASK_CLEANUP_DAYS=7

# --- Redis settings ---
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=password12345
REDIS_DB=0
REDIS_CACHE_TTL=30

# --- Logging settings ---
LOG_LEVEL=DEBUG
LOG_FILE=logs/app.log
ENVIRONMENT=development
```

3. Запустите приложение:
```bash
go mod download
go run cmd/app/main.go
```

### Запуск через Docker
1. Убедитесь, что у вас установлен Docker и Docker Compose.
2. Запустите сервисы:
```bash
docker-compose up --build
```

### Структура проекта
```
.
├── api/                 # API документация
├── cmd/                 # Точки входа приложения
├── docker/             # Dockerfile и docker-compose
├── internal/           # Внутренняя логика приложения
│   ├── config/         # Конфигурация приложения
│   ├── delivery/       # Обработчики HTTP запросов
│   ├── domain/         # Модели данных
│   ├── logger/         # Настройка логгера
│   ├── handler/        # Обработчики бизнес-логики
│   ├── repository/     # Работа с БД и Redis
│   └── useCase/        # Бизнес-логика
├── migrations/         # Миграции базы данных
├── pkg/                # Общие утилиты
└── logs/               # Логи приложения
```

## Тестирование API через Postman

### P.S Swagger UI доступен при запуске приложения по:

``
http://localhost:8085/swagger/index.html#/
``

### 1. Регистрация пользователя
**POST** `/auth/register`

1. Откройте Postman.
2. Выберите метод `POST`.
3. Введите URL: `http://localhost:8085/auth/register`.
4. Перейдите на вкладку **Body** и выберите **raw**.
5. Вставьте следующий JSON:

```json
{
    "username": "john_doe",
    "email": "john.doe@example.com",
    "password": "password123"
}
```

6. Нажмите **Send**.

**Ответ:**

```json
{
  "message": "Пользователь успешно зарегистрирован"
}
```

### 2. Авторизация пользователя (получение JWT токена)
**POST** `/auth/login`

1. Откройте Postman.
2. Выберите метод `POST`.
3. Введите URL: `http://localhost:8085/auth/login`.
4. Перейдите на вкладку **Body** и выберите **raw**.
5. Вставьте следующий JSON:

```json
{
    "email": "john.doe@example.com",
    "password": "password123"
}
```

6. Нажмите **Send**.

**Ответ:**

```json
{
  "access_token": "your_access_token",
  "refresh_token": "your_refresh_token"
}
```

### 3. Создание задачи
**POST** `/tasks`

1. Откройте Postman.
2. Выберите метод `POST`.
3. Введите URL: `http://localhost:8085/tasks`.
4. Перейдите на вкладку **Body** и выберите **raw**.
5. Вставьте следующий JSON:

```json
{
  "title": "Task 1",
  "description": "Task description",
  "status": "pending",
  "priority": "high",
  "due_date": "2025-04-21T10:00:00Z"
}
```

6. Нажмите **Send**.

**Ответ:**

```json
{
  "id": 1,
  "title": "Task 1",
  "description": "Task description",
  "status": "pending",
  "priority": "high",
  "due_date": "2025-04-21T10:00:00Z",
  "created_at": "2025-04-19T17:15:13.4321104+03:00",
  "updated_at": "2025-04-19T17:15:13.4321104+03:00"
}
```

### 4. Получение всех задач без фильтров
**GET** `/tasks`

1. Откройте Postman.
2. Выберите метод `GET`.
3. Введите URL: `http://localhost:8085/tasks`.
4. Нажмите **Send**.

**Ответ:**

```json
[
  {
    "id": 2,
    "title": "Task 2",
    "description": "Task description 2",
    "status": "pending",
    "priority": "high",
    "due_date": "2025-04-21T10:00:00Z",
    "created_at": "2025-04-19T14:15:13.43211Z",
    "updated_at": "2025-04-19T14:15:13.43211Z"
  },
  {
    "id": 3,
    "title": "Задача 3",
    "description": "Описание задачи 3",
    "status": "pending",
    "priority": "high",
    "due_date": "2025-05-01T00:00:00Z",
    "created_at": "2025-04-18T14:11:46.027575Z",
    "updated_at": "2025-04-18T14:11:46.027575Z"
  },
  {
    "...": "..."
  }
]
```

### 5. Обновление задачи с возможностью частичного изменения данных
**PUT** `/tasks/:id`

1. Откройте Postman.
2. Выберите метод `PUT`.
3. Введите URL: `http://localhost:8085/tasks/:id`.
4. Перейдите на вкладку **Body** и выберите **raw**.
5. Вставьте следующий JSON:

```json
{
  "status": "done",
  "priority": "low"
}
```

6. Нажмите **Send**.

**Ответ:**

```json
{
  "id": 1,
  "status": "done",
  "priority": "low",
  "due_date": "0001-01-01T00:00:00Z",
  "created_at": "0001-01-01T00:00:00Z",
  "updated_at": "2025-04-19T17:22:58.7867265+03:00"
}
```

### 6. Удаление зачачи
**DELETE** `/tasks/:id`

1. Откройте Postman.
2. Выберите метод `DELETE`.
3. Введите URL: `http://localhost:8085/tasks/:id`.
4. За место `:id` вставте id задачи, которую надо удалить
6. Нажмите **Send**.

### 7. Получение аналитики за неделю
**GET** `/tasks/:id`

1. Откройте Postman.
2. Выберите метод `GET`.
3. Введите URL: `http://localhost:8085/analytics`.
4. Нажмите **Send**.

**Ответ:**

```json
{
  "status_counts": {
    "done": 14,
    "in_progress": 7,
    "pending": 13
  },
  "average_execution_time": "170 часов 2 минут 0 секунд",
  "report_last_period": {
    "completed_tasks": 14,
    "overdue_tasks": 0
  }
}
```

### 8. Импорт задач из JSON
**POST** `/tasks/import`

1. Откройте Postman.
2. Выберите метод `POST`.
3. Введите URL: `http://localhost:8085/tasks/import`.
4. Перейдите на вкладку **Body** и выберите **form-data**.
5. Добавьте новый ключ со значенем:
- **Key**: `file`
- **Value**: `exported file`
6. Нажмите **Send**.

**Ответ:**

```json
{
  "inserted_tasks": 3,
  "message": "Импорт успешно завершен",
  "skipped_tasks": [
    "Задача 3: приоритет задачи не может быть пустым",
    "Задача 5: название задачи не может быть пустым",
    "Задача 4: некорректный статус задачи: "
  ]
}
```

### 9. Экспорт задач в JSON
**GET** `/tasks/export`

1. Откройте Postman.
2. Выберите метод `GET`.
3. Введите URL: `http://localhost:8085/tasks/export`.
4. Нажмите **Send**.

## Пример файла JSON для импорта/экспорта задач

Пример файла для импорта задач в формате JSON:
```json
[
  {
    "title": "Задача 1",
    "description": "Описание задачи 1",
    "status": "pending",
    "priority": "high",
    "due_date": "2025-05-01T00:00:00Z"
  },
  {
    "title": "Задача 2",
    "description": "Описание задачи 2",
    "status": "in_progress",
    "priority": "medium",
    "due_date": "2025-05-02T00:00:00Z"
  }
]
```

## Миграции
Чтобы выполнить миграции для создания таблиц в базе данных, выполните следующие шаги:

1. Убедитесь, что PostgreSQL запущен.
2. Примените миграции:
```bash
migrate -path=./migrations -database="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up
```

## Тестирование
Для запуска тестов:
```bash
go test ./...
```

## Миграции базы данных
Миграции выполняются автоматически при запуске приложения через Docker Compose, с использованием пакета [migrate](https://github.com/golang-migrate/migrate).

Скрипты миграций находятся в папке `migrations`:
1. `001_create_users.up.sql` — создание таблицы пользователей.
2. `002_create_tasks.up.sql` — создание таблицы задач.

### Запуск миграций вручную
Если необходимо вручную запустить миграции, используйте команду:
```bash
docker-compose exec app migrate -path /migrations -database postgres://postgres:postgres@postgres:5432/go_tasker?sslmode=disable up
```

## Мониторинг
Система включает базовый мониторинг:
- Логирование ошибок и важных событий

## Из необходимых минимальных доработок
1. Правильная обработка и анализ ошибок
2. Всего по чуть-чуть)