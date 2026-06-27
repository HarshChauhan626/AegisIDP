.PHONY: dev build test migrate lint backend-dev frontend-dev tidy

## dev: Start all services via Docker Compose
dev:
	docker-compose up --build

## backend-dev: Run backend locally (without Docker)
backend-dev:
	cd backend && go run ./cmd/main.go

## frontend-dev: Run frontend locally (without Docker)
frontend-dev:
	cd frontend && npm run dev

## build: Build backend binary + frontend production bundle
build: build-backend build-frontend

build-backend:
	cd backend && go build -o ../bin/platform-orchestrator ./cmd/main.go

build-frontend:
	cd frontend && npm run build

## test: Run backend unit tests
test:
	cd backend && go test ./... -v -race

## migrate: Run DB migrations (auto-migrate via GORM on startup)
migrate:
	cd backend && go run ./cmd/main.go --migrate-only

## lint: Lint backend + frontend
lint: lint-backend lint-frontend

lint-backend:
	cd backend && go vet ./...

lint-frontend:
	cd frontend && npm run lint

## tidy: Tidy go modules
tidy:
	cd backend && go mod tidy

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'
