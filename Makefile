APP_PACKAGE := ./cmd/api
APP_BINARY := bin/api
GOLANGCI_LINT_VERSION := v2.12.2
GOIMPORTS_PACKAGE := golang.org/x/tools/cmd/goimports@latest
ifneq (,$(wildcard .env))
include .env
export
endif
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/workout_tracker?sslmode=disable

.DEFAULT_GOAL := help

.PHONY: help run build test test-cover fmt fmt-check lint check tidy swagger clean env install-hooks docker-up docker-up-db docker-down docker-logs migrate-up migrate-down migrate-force seed seed-docker

help: ## Показать список доступных make-команд
	@echo "Доступные команды:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z0-9_-]+:.*?##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf "\n"

run: ## Запустить API локально через go run
	go run $(APP_PACKAGE)

build: ## Собрать бинарник API в bin/api
	go build -o $(APP_BINARY) $(APP_PACKAGE)

test: ## Запустить все тесты
	go test ./...

test-cover: ## Запустить тесты с отчетом покрытия
	go test -cover ./...

fmt: ## Отформатировать Go-код и оптимизировать импорты через goimports
	go run $(GOIMPORTS_PACKAGE) -w cmd internal docs

fmt-check: ## Проверить форматирование и импорты без изменения файлов
	@test -z "$$(go run $(GOIMPORTS_PACKAGE) -l cmd internal docs)"

lint: ## Запустить Go-линтеры через golangci-lint в Docker
	docker run --rm -v "$$(pwd):/app" -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run

check: fmt-check lint test ## Запустить форматирование в check-режиме, линтеры и тесты

tidy: ## Обновить go.mod и go.sum
	go mod tidy

swagger: ## Сгенерировать Swagger-документацию в папку docs
	go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -d cmd/api,internal/http/handlers,internal/domain

clean: ## Удалить локальные артефакты сборки
	rm -rf bin

env: ## Создать .env из .env.example, если файла еще нет
	cp -n .env.example .env

install-hooks: ## Установить git hooks из scripts/git-hooks
	@test -d .git || (echo "Git-репозиторий не найден. Сначала выполните: git init" && exit 1)
	cp scripts/git-hooks/pre-push .git/hooks/pre-push
	chmod +x .git/hooks/pre-push
	@echo "Git hook pre-push установлен"

docker-up: ## Собрать и запустить API вместе с PostgreSQL
	docker compose up --build

docker-up-db: ## Запустить только PostgreSQL в Docker
	docker compose up -d postgres

docker-down: ## Остановить контейнеры Docker Compose
	docker compose down

docker-logs: ## Показать логи контейнеров Docker Compose
	docker compose logs -f

migrate-up: ## Применить все новые миграции локально через golang-migrate
	migrate -path migrations -database "$(DATABASE_URL)" up

seed: ## Загрузить тестовые данные в локальную PostgreSQL через psql
	psql "$(DATABASE_URL)" -f seeds/dev.sql

seed-docker: ## Загрузить тестовые данные через Docker Compose
	docker compose --profile tools run --rm seed

migrate-down: ## Откатить последнюю миграцию локально через golang-migrate
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-force: ## Принудительно выставить версию миграции: make migrate-force VERSION=1
ifndef VERSION
	$(error Укажите VERSION, например: make migrate-force VERSION=1)
endif
	migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)
