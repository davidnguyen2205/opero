SHELL := /bin/bash

.PHONY: help generate oapi sqlc migrate lint vet fmt fmt-check test run tidy up down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

generate: oapi sqlc ## Regenerate all code from the spec and SQL

oapi: ## Generate server types/interfaces from api/openapi.yaml
	oapi-codegen -config oapi-codegen.yaml ../api/openapi.yaml

sqlc: ## Generate type-safe DB code (needs migrations + queries; populated from M1)
	sqlc generate

migrate: ## Run migrations (control-plane + fan-out to tenants); implemented in M1
	go run ./cmd/migrate

lint: ## Run golangci-lint
	golangci-lint run

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go code in place
	gofmt -w .

fmt-check: ## Fail if any Go file is not gofmt-clean
	@test -z "$$(gofmt -l .)" || { echo "gofmt needed:"; gofmt -l .; exit 1; }

test: ## Run all tests
	go test ./...

run: ## Run the API server (serves /health)
	go run ./cmd/api

tidy: ## Resolve and pin module dependencies
	go mod tidy

up: ## Start local Postgres (control-plane DB)
	docker compose up -d

down: ## Stop local Postgres
	docker compose down
