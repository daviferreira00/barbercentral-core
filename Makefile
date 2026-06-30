.PHONY: help run build tidy test docker-build docker-up docker-down docker-reset migrate-up migrate-down migrate-create

APP_NAME       := barbercentral-backend
MAIN_PATH      := ./cmd/api
MIGRATIONS_DIR := migrations
GOBIN          := $(shell go env GOBIN)
GOPATH         := $(shell go env GOPATH)
MIGRATE_BINARY := $(or $(shell command -v migrate 2>/dev/null),$(if $(GOBIN),$(GOBIN)/migrate,$(GOPATH)/bin/migrate))

# Carrega o .env para os comandos de migration.
ifneq (,$(wildcard .env))
include .env
export
endif

# URL de conexao no formato exigido pelo golang-migrate (MySQL).
DB_PORT := $(or $(DB_PORT),3306)
DB_URL := mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)

help: ## Lista os comandos disponiveis
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

run: ## Roda a aplicacao localmente (go run)
	@go run $(MAIN_PATH)

build: ## Compila o binario em bin/
	@go build -ldflags="-w -s" -o bin/$(APP_NAME) $(MAIN_PATH)

tidy: ## Organiza as dependencias (go mod tidy)
	@go mod tidy

test: ## Roda os testes
	@go test -race ./...

docker-build: ## Builda a imagem Docker da aplicacao
	@docker compose build

docker-up: ## Sobe a aplicacao em container
	@docker compose up -d

docker-down: ## Derruba os containers da aplicacao
	@docker compose down

docker-reset: ## Para, compila e sobe os containers novamente
	@docker compose down
	@docker compose build
	@docker compose up -d

install-migrate:
	@echo "Instalando golang-migrate..."
	@go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

migrate-up: ## Aplica as migrations pendentes
	@echo "Usando migrate em: $(MIGRATE_BINARY)"
	@if [ ! -x "$(MIGRATE_BINARY)" ]; then \
	  echo "Erro: migrate não encontrado. Rode 'make install-migrate' ou adicione $(GOPATH)/bin ao PATH."; exit 1; \
	fi
	@$(MIGRATE_BINARY) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down: ## Reverte a ultima migration
	@echo "Usando migrate em: $(MIGRATE_BINARY)"
	@if [ ! -x "$(MIGRATE_BINARY)" ]; then \
	  echo "Erro: migrate não encontrado. Rode 'make install-migrate' ou adicione $(GOPATH)/bin ao PATH."; exit 1; \
	fi
	@$(MIGRATE_BINARY) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-create: ## Cria uma migration nova: make migrate-create name=nome_da_migration
	@echo "Usando migrate em: $(MIGRATE_BINARY)"
	@if [ ! -x "$(MIGRATE_BINARY)" ]; then \
	  echo "Erro: migrate não encontrado. Rode 'make install-migrate' ou adicione $(GOPATH)/bin ao PATH."; exit 1; \
	fi
	@$(MIGRATE_BINARY) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
