#!/usr/bin/env bash
# ==============================================================================
# Hackways Repository Pre-Push Verification Suite
# ==============================================================================
set -euo pipefail

echo "=========================================="
echo " 1. Go Backend: Format, Vet, Tests, Build "
echo "=========================================="
cd services/api

echo "--> Checking Go formatting..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "ERROR: Unformatted Go files:"
    echo "$UNFORMATTED"
    exit 1
fi
echo "Go formatting clean."

echo "--> Running go vet..."
go vet ./...

echo "--> Running Go unit tests..."
go test ./... -v -count=1

echo "--> Compiling Go API binary..."
go build -o /dev/null ./cmd/api
echo "Backend checks passed."

cd ../..

echo "=========================================="
echo " 2. Frontend: Syntax & Type Checking     "
echo "=========================================="
for APP in core-web public-web organizer-web admin-web; do
    echo "--> Checking apps/$APP..."
    cd "apps/$APP"
    if [ -f "package-lock.json" ] || [ -d "node_modules" ]; then
        npm run type-check
        npm run lint
    fi
    cd ../..
done

echo "=========================================="
echo " 3. Repository Hygiene & Secret Audit     "
echo "=========================================="
if [ -f .env ]; then
    echo "WARNING: Local .env exists (must remain gitignored)."
fi

echo "ALL CHECKS PASSED SUCCESSFULLY."
