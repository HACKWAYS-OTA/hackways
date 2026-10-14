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

echo "==> Starting Core Backend (port 8081)..."
(cd services/core-backend && go run ./cmd/server) &
CORE_PID=$!

echo "==> Starting Global Backend (port 8082)..."
(cd services/global-backend && go run ./cmd/server) &
GLOBAL_PID=$!

echo "==> Starting Core Web frontend (port 3001)..."
(cd apps/core-web && npm run dev) &
WEB_PID=$!

trap "kill $CORE_PID $GLOBAL_PID $WEB_PID 2>/dev/null || true; exit" INT TERM
wait
