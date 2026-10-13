.PHONY: help dev dev-api dev-web db-up db-down db-logs migrate-up migrate-down test test-api test-web lint format build clean

SHELL := /bin/bash

# Default target
all: help

## help: Display available Makefile commands
help:
	@echo "Hackways Platform — Available Commands:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## dev: Start database containers and launch development servers
dev: db-up
	@echo "Starting local development services..."
	@make -j 2 dev-api dev-web

## dev-api: Run Go API backend server
dev-api:
	cd services/api && go run ./cmd/api

## dev-web: Run Next.js frontend application (core-web on port 3001)
dev-web:
	cd apps/core-web && npm run dev

## db-up: Start local PostgreSQL and Mailpit containers in background
db-up:
	docker compose up -d

## db-down: Stop local database and mail services
db-down:
	docker compose down

## db-logs: Stream logs from database containers
db-logs:
	docker compose logs -f

## migrate-up: Apply all pending database schema migrations
migrate-up:
	migrate -path services/api/migrations -database "$$DATABASE_URL" up

## migrate-down: Rollback the most recent database migration
migrate-down:
	migrate -path services/api/migrations -database "$$DATABASE_URL" down 1

## test: Run all automated tests across backend and frontend
test: test-api test-web

## test-api: Run Go backend tests with race detector enabled
test-api:
	cd services/api && go test ./... -v -race -count=1

## test-web: Run Next.js frontend test suite across all apps
test-web:
	cd apps/core-web && npm test
	cd apps/public-web && npm test
	cd apps/organizer-web && npm test
	cd apps/admin-web && npm test

## lint: Run static analysis and lint checks across the repository
lint:
	cd services/api && go vet ./...
	cd apps/core-web && npm run lint && npm run type-check
	cd apps/public-web && npm run lint && npm run type-check
	cd apps/organizer-web && npm run lint && npm run type-check
	cd apps/admin-web && npm run lint && npm run type-check

## format: Format source code across Go and TypeScript files
format:
	cd services/api && gofmt -w ./...
	cd apps/core-web && npm run format 2>/dev/null || true
	cd apps/public-web && npm run format 2>/dev/null || true
	cd apps/organizer-web && npm run format 2>/dev/null || true
	cd apps/admin-web && npm run format 2>/dev/null || true

## build: Compile production Go binary and Next.js frontend applications
build:
	cd services/api && CGO_ENABLED=0 go build -o bin/api ./cmd/api
	cd apps/core-web && npm run build
	cd apps/public-web && npm run build
	cd apps/organizer-web && npm run build
	cd apps/admin-web && npm run build

## clean: Remove compiled binaries and build artifacts
clean:
	rm -rf services/api/bin/
	rm -rf apps/core-web/.next/ apps/public-web/.next/ apps/organizer-web/.next/ apps/admin-web/.next/
	rm -rf apps/core-web/out/ apps/public-web/out/ apps/organizer-web/out/ apps/admin-web/out/
	rm -f services/api/coverage.out
