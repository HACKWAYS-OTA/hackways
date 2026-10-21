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
	@make -j 3 dev-core dev-global dev-web

## dev-core: Run Go Core Backend engine (:8081)
dev-core:
	cd services/core-backend && go run ./cmd/server

## dev-global: Run Go Global Backend engine (:8082)
dev-global:
	cd services/global-backend && go run ./cmd/server

## dev-web: Run Next.js frontend application (core-web on port 3001)
dev-web:
	cd apps/core-web && npm run dev

## db-up: Start the local PostgreSQL container
db-up:
	docker compose up -d

## db-down: Stop the local database service
db-down:
	docker compose down

## db-logs: Stream logs from database containers
db-logs:
	docker compose logs -f

## migrate-up: Apply all pending database schema migrations
migrate-up:
	migrate -path deploy/migrations -database "$$DATABASE_URL" up

## migrate-down: Rollback the most recent database migration
migrate-down:
	migrate -path deploy/migrations -database "$$DATABASE_URL" down 1

## test: Run all automated tests across backend and frontend
test: test-api test-web

## test-api: Run Go backend tests with race detector enabled
test-api:
	cd services/core-backend && go test ./... -v -race -count=1
	cd services/global-backend && go test ./... -v -race -count=1

## test-web: Run Next.js frontend test suite across all apps
test-web:
	cd apps/core-web && npm test
	cd apps/public-web && npm test
	cd apps/organizer-web && npm test
	cd apps/admin-web && npm test

## lint: Run static analysis and lint checks across the repository
lint:
	cd services/core-backend && go vet ./...
	cd services/global-backend && go vet ./...
	cd apps/core-web && npm run lint && npm run type-check
	cd apps/public-web && npm run lint && npm run type-check
	cd apps/organizer-web && npm run lint && npm run type-check
	cd apps/admin-web && npm run lint && npm run type-check

## format: Format source code across Go and TypeScript files
format:
	cd services/core-backend && gofmt -w .
	cd services/global-backend && gofmt -w .
	cd apps/core-web && npm run format 2>/dev/null || true
	cd apps/public-web && npm run format 2>/dev/null || true
	cd apps/organizer-web && npm run format 2>/dev/null || true
	cd apps/admin-web && npm run format 2>/dev/null || true

## build: Compile production Go binaries and Next.js frontend applications
build:
	cd services/core-backend && CGO_ENABLED=0 go build -o bin/core-backend ./cmd/server
	cd services/global-backend && CGO_ENABLED=0 go build -o bin/global-backend ./cmd/server
	cd apps/core-web && npm run build
	cd apps/public-web && npm run build
	cd apps/organizer-web && npm run build
	cd apps/admin-web && npm run build

## clean: Remove compiled binaries and build artifacts
clean:
	rm -rf services/core-backend/bin/ services/global-backend/bin/ services/api/bin/
	rm -rf apps/core-web/.next/ apps/public-web/.next/ apps/organizer-web/.next/ apps/admin-web/.next/
	rm -rf apps/core-web/out/ apps/public-web/out/ apps/organizer-web/out/ apps/admin-web/out/
	rm -f services/core-backend/coverage.out services/global-backend/coverage.out services/api/coverage.out
