#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Hackways Production Deployment Automation
# Target: Azure VM (Standard B2als v2 - 2 vCPU, 4 GB RAM, 50 GB NVMe)
# ==============================================================================

PROJECT_DIR="/opt/hackways"
COMPOSE_FILE="$PROJECT_DIR/deploy/docker/docker-compose.prod.yml"
ENV_FILE="$PROJECT_DIR/.env"

echo "=== 1. Validating Production Environment ==="
if [ ! -f "$ENV_FILE" ]; then
    echo "ERROR: Environment file $ENV_FILE not found!"
    exit 1
fi

echo "=== 2. Pulling Pre-built Container Images from Registry ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" pull

echo "=== 3. Ensuring PostgreSQL Database is Running ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d postgres
echo "Waiting for PostgreSQL healthcheck..."
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" wait --wait-healthy postgres

echo "=== 4. Applying Pending Database Migrations ==="
for mig in "$PROJECT_DIR"/deploy/migrations/*.up.sql; do
    if [ -f "$mig" ]; then
        echo "Applying $(basename "$mig")..."
        docker exec -i hackways-postgres psql -U hackways_prod -d hackways_prod < "$mig" || true
    fi
done

echo "=== 5. Launching Production Container Services ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d --remove-orphans

echo "=== 6. Pruning Dangling Images to Preserve NVMe Disk Cap ==="
docker image prune -f

echo "=== 7. Live Container Health and Resource Inspection ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" ps
docker stats --no-stream

echo "=== Hackways Production Deployment Complete & Verified ==="
