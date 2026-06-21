# Workout Tracker API

Backend API для учета выполненных упражнений и получения статистики по пользователю.

## Требования

- Docker и Docker Compose
- Для запуска через Docker Compose достаточно Docker и Docker Compose
- Для локального запуска без Docker дополнительно нужен Go 1.25+
- `goimports` в `PATH` нужен только для `make fmt` и `make check`
- `golang-migrate` CLI нужен только для локального запуска миграций через `make migrate-up`

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

## Контракты API

### `POST /exercises`

Создает упражнение.

Запрос:

```json
{ "name": "Bench Press" }
```

Ответ `201 Created`:

```json
{
  "id": 1,
  "name": "Bench Press",
  "created_at": "2026-06-21T10:00:00Z"
}
```

Коды ошибок:

- `400 Bad Request` - невалидный JSON или неизвестные поля
- `422 Unprocessable Entity` - ошибка валидации, например пустое или неочищенное имя
- `409 Conflict` - упражнение с таким именем уже существует
- `500 Internal Server Error` - ошибка БД или сервиса

### `POST /executions`

Фиксирует выполнение упражнения.

Запрос:

```json
{
  "user_id": "user-1",
  "exercise_id": 1,
  "performed_at": "2026-06-21T10:00:00Z"
}
```

Поле `performed_at` необязательно. Если не передано, используется текущее время.

Ответ `201 Created`:

```json
{
  "id": 1,
  "user_id": "user-1",
  "exercise_id": 1,
  "performed_at": "2026-06-21T10:00:00Z",
  "created_at": "2026-06-21T10:00:01Z"
}
```

Коды ошибок:

- `400 Bad Request` - невалидный JSON или неизвестные поля
- `422 Unprocessable Entity` - ошибка валидации, например некорректный `user_id`, `exercise_id <= 0` или `performed_at` в будущем
- `404 Not Found` - упражнение не найдено
- `500 Internal Server Error` - ошибка БД или сервиса

### `GET /users/{userID}/statistics`

Возвращает агрегированную статистику по пользователю.

Ответ `200 OK`:

```json
{
  "user_id": "user-1",
  "total": 10,
  "today": 2,
  "last_7_days": [
    { "date": "2026-06-15", "count": 0 },
    { "date": "2026-06-16", "count": 1 },
    { "date": "2026-06-17", "count": 0 },
    { "date": "2026-06-18", "count": 0 },
    { "date": "2026-06-19", "count": 3 },
    { "date": "2026-06-20", "count": 4 },
    { "date": "2026-06-21", "count": 2 }
  ]
}
```

Коды ошибок:

- `422 Unprocessable Entity` - некорректный `userID` в path
- `500 Internal Server Error` - ошибка БД или сервиса

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
- `cmd/api` - точка входа приложения
- `internal/config` - конфигурация из окружения
- `internal/database` - подключение к PostgreSQL
- `internal/domain` - доменные модели
- `internal/http`
  - `handlers` - HTTP-обработчики, валидация и ответы
  - `middleware` - middleware для логирования запросов
  - `router.go` - сборка маршрутов
- `internal/repository` - SQL-репозитории для PostgreSQL
- `internal/service` - бизнес-слой
- `migrations` - SQL-миграции `golang-migrate`
- `seeds` - тестовые данные для локальной разработки
- `docs` - Swagger/OpenAPI-спецификация
- `Dockerfile`, `docker-compose.yml` - контейнеризация и запуск приложения
- `Makefile` - команды для локальной разработки и запуска

## Архитектура

- Слоистая архитектура: HTTP-слой, сервисный слой и слой доступа к данным разделены по ответственности.
- Repository pattern используется для инкапсуляции SQL и работы с PostgreSQL.
- Service layer держит бизнес-правила, валидацию и преобразование данных между HTTP и репозиториями.
- Dependency injection используется через конструкторы, чтобы упростить тестирование и замену реализаций.
- API выполнен в REST-стиле и остается stateless: состояние не хранится в приложении между запросами.
- Статистика считается одним SQL-запросом в PostgreSQL, чтобы не тащить агрегации в приложение и не делать лишние round-trip.

## Технические решения

- `chi` используется для маршрутизации.
- `pgx` используется для подключения к PostgreSQL.
- `golang-migrate` используется для версионирования схемы базы данных.
- `validator` используется для базовой проверки входных JSON-запросов.
- `slog` используется для структурированных JSON-логов.
- Агрегации статистики выполняются на стороне PostgreSQL.
- Пользователь в этой задаче моделируется внешним строковым `user_id`, а не отдельной таблицей пользователей. Это упрощает API и соответствует требованиям задания, где аутентификация не нужна.
- Поле `today` считается относительно текущей даты PostgreSQL через `CURRENT_DATE`. В `docker-compose.yml` таймзона PostgreSQL зафиксирована на `UTC`, поэтому границы суток не зависят от окружения.
- Статистика считается одним запросом к PostgreSQL, чтобы избежать лишних round-trip в приложение и сохранить логику агрегации рядом с данными.
