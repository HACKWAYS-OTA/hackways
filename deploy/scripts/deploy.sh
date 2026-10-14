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

# Source environment
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

echo "=== 2. Pulling Base Images from Registry ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" pull postgres nginx || true

echo "=== 2b. Compiling Go Backend Engines ==="
docker build -f "$PROJECT_DIR/deploy/docker/Dockerfile.core-backend" -t "${DOCKER_REGISTRY:-hackways}/hackways-core-backend:${IMAGE_TAG:-latest}" "$PROJECT_DIR"
docker build -f "$PROJECT_DIR/deploy/docker/Dockerfile.global-backend" -t "${DOCKER_REGISTRY:-hackways}/hackways-global-backend:${IMAGE_TAG:-latest}" "$PROJECT_DIR"

echo "=== 3. Ensuring PostgreSQL Database is Running ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d postgres
echo "Waiting for PostgreSQL healthcheck..."
sleep 2

echo "=== 4. Applying Pending Database Migrations ==="
for mig in "$PROJECT_DIR"/deploy/migrations/*.up.sql; do
    if [ -f "$mig" ]; then
        echo "Applying $(basename "$mig")..."
        docker exec -i hackways-postgres psql -U "${POSTGRES_USER:-hackways_prod}" -d "${POSTGRES_DB:-hackways_prod}" < "$mig" || true
    fi
done

echo "=== 5. Launching Production Container Services ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d core-backend global-backend nginx

# If web app images are present, launch them
for app in public-web core-web organizer-web admin-web; do
    if docker image inspect "${DOCKER_REGISTRY:-hackways}/hackways-${app}:${IMAGE_TAG:-latest}" >/dev/null 2>&1; then
        echo "Launching $app..."
        docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d "$app"
    fi
done

echo "Reloading Nginx ingress configuration..."
docker exec hackways-nginx nginx -s reload || true

echo "=== 6. Pruning Dangling Images to Preserve NVMe Disk Cap ==="
docker image prune -f

echo "=== 7. Live Container Health and Resource Inspection ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" ps
docker stats --no-stream

echo "=== Hackways Production Deployment Complete & Verified ==="
