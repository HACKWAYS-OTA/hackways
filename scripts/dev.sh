#!/usr/bin/env bash
# ==============================================================================
# Hackways Local Development Launcher
# ==============================================================================
set -euo pipefail

echo "==> Checking prerequisites..."
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed."; exit 1; }
command -v go >/dev/null 2>&1 || { echo "Go 1.23+ is required but not installed."; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Node.js 22+ is required but not installed."; exit 1; }

echo "==> Starting PostgreSQL and Mailpit via Docker Compose..."
docker compose up -d

echo "==> Starting Go API server..."
(cd services/api && go run ./cmd/api) &
API_PID=$!

echo "==> Starting Next.js frontend..."
(cd apps/web && npm run dev) &
WEB_PID=$!

trap "kill $API_PID $WEB_PID 2>/dev/null || true; exit" INT TERM
wait
