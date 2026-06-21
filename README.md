# Workout Tracker API

Backend API для учета выполненных упражнений и получения статистики по пользователю.

## Требования

- Go 1.25+
- Docker и Docker Compose
- golang-migrate CLI, если миграции запускаются локально через `make migrate-up`

## Запуск

```bash
docker compose up -d postgres
docker compose --profile tools run --rm migrate
make docker-up
```

Для пересборки образа приложения:

```bash
make docker-rebuild
```

## Проверки кода

```bash
make fmt
make lint
make check
make test-cover
make test-integration
```

`make fmt` форматирует Go-код и оптимизирует импорты через `goimports`.
`make lint` запускает `golangci-lint` в Docker.
`make check` проверяет форматирование, запускает линтеры и тесты.
`make test-integration` прогоняет интеграционные тесты репозитория на PostgreSQL. Для них нужен `TEST_DATABASE_URL` или `DATABASE_URL`.

Установить git hook, который запускает `make check` перед `git push`:

```bash
make install-hooks
```

Локальный запуск без Docker для приложения:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/workout_tracker?sslmode=disable"
make run
```

## Миграции

Через Docker:

```bash
docker compose --profile tools run --rm migrate
```

Загрузить тестовые данные:

```bash
make seed-docker
```

Локально:

```bash
make migrate-up
make seed
make migrate-down
```

## API

Swagger UI доступен после запуска приложения:

```text
http://localhost:3000/swagger/index.html
```

Сгенерировать Swagger-документацию:

```bash
make swagger
```

Health check:

```bash
curl http://localhost:3000/health
```

Создать упражнение:

```bash
curl -X POST http://localhost:3000/exercises \
  -H "Content-Type: application/json" \
  -d '{"name":"Bench Press"}'
```

Зафиксировать выполнение:

```bash
curl -X POST http://localhost:3000/executions \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user-1","exercise_id":1,"performed_at":"2026-06-21T10:00:00Z"}'
```

Получить статистику:

```bash
curl http://localhost:3000/users/user-1/statistics
```

После применения seed можно проверить демо-пользователей:

```bash
curl http://localhost:3000/users/demo-user-1/statistics
curl http://localhost:3000/users/demo-user-2/statistics
```

## Структура

```text
cmd/api              точка входа приложения
internal/config      конфигурация из окружения
internal/database    подключение к PostgreSQL
internal/domain      доменные модели
internal/http        router, handlers, middleware
internal/repository  работа с PostgreSQL
internal/service     бизнес-слой
migrations           SQL-миграции golang-migrate
seeds                тестовые данные для локальной разработки
```

## Технические решения

- `chi` используется для маршрутизации.
- `pgx` используется для подключения к PostgreSQL.
- `golang-migrate` используется для версионирования схемы базы данных.
- `validator` используется для базовой проверки входных JSON-запросов.
- `slog` используется для структурированных JSON-логов.
- Агрегации статистики выполняются на стороне PostgreSQL.
