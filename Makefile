.PHONY: help build test lint tidy proto proto-tools run-panel run-node docker-up docker-down

BIN := bin
GOFLAGS :=
VERSION := $(shell cat VERSION 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}'

build: ## Build panel and node binaries
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN)/panel ./cmd/panel
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN)/node ./cmd/node

certs: ## Generate a dev mTLS chain into deploy/certs
	go run ./cmd/gencerts -out deploy/certs -san localhost,127.0.0.1

test: ## Run all tests with race detector
	go test -race -count=1 ./...

cover: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

lint: ## Run golangci-lint (install: https://golangci-lint.run)
	golangci-lint run ./...

tidy: ## Sync go.mod/go.sum
	go mod tidy

proto-tools: ## Install buf + protoc Go plugins
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto: ## Lint and regenerate gRPC code from proto/
	buf lint proto
	buf generate proto

sqlc: ## Regenerate type-safe DB code from SQL
	sqlc generate

db-tools: ## Install sqlc + goose
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

migrate-up: ## Apply migrations (needs VORTEX_DATABASE_URL)
	goose -dir migrations postgres "$$VORTEX_DATABASE_URL" up

migrate-down: ## Roll back the last migration
	goose -dir migrations postgres "$$VORTEX_DATABASE_URL" down

test-db: ## Run integration tests against a throwaway DB (needs docker compose up)
	VORTEX_TEST_DB="postgres://vortex:vortex@localhost:5432/vortex?sslmode=disable" go test -count=1 ./internal/platform/postgres/...

run-panel: ## Run the control plane
	go run ./cmd/panel

run-node: ## Run a node agent
	go run ./cmd/node

docker-up: ## Start dev dependencies (postgres+timescale, redis)
	docker compose up -d

docker-down: ## Stop dev dependencies
	docker compose down

images: ## Build production panel + node images
	docker build -f deploy/Dockerfile --target panel -t vortexui/panel .
	docker build -f deploy/Dockerfile --target node  -t vortexui/node  .

stack-up: ## Build & run the full stack (needs `make certs` and JWT_SECRET)
	docker compose -f deploy/compose.yml up --build -d

stack-down: ## Tear down the full stack
	docker compose -f deploy/compose.yml down

docs: ## Build public docs site (Arena → review/site)
	python review/patch_arena.py

docs-wiki: ## Build MkDocs wiki (reference)
	pip install -r docs/requirements.txt
	mkdocs build

docs-serve: ## Serve Arena docs locally (build first)
	python review/patch_arena.py
	python -m http.server 8000 --directory review/site

# --- SDK Generation ---

SDK_OUT := sdk

sdk-tools: ## Install openapi-generator CLI
	npm install -g @openapitools/openapi-generator-cli

sdk-python: ## Generate Python client SDK from OpenAPI spec
	openapi-generator-cli generate -i docs/openapi.yaml -g python -o $(SDK_OUT)/python --additional-properties=packageName=vortexui_client

sdk-javascript: ## Generate JavaScript/TypeScript client SDK from OpenAPI spec
	openapi-generator-cli generate -i docs/openapi.yaml -g typescript-axios -o $(SDK_OUT)/javascript --additional-properties=npmName=vortexui-client

sdk-go: ## Generate Go client SDK from OpenAPI spec
	openapi-generator-cli generate -i docs/openapi.yaml -g go -o $(SDK_OUT)/go --additional-properties=packageName=vortexui

sdk-all: sdk-python sdk-javascript sdk-go ## Generate all client SDKs

openapi: ## Regenerate OpenAPI spec from source annotations
	go run github.com/swaggo/swag/cmd/swag@latest init -g internal/panel/api/server.go -o docs --outputTypes yaml --parseDependency --parseInternal
