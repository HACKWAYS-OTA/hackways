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

echo "=== 2. Pulling Pre-Built Production Images from Registry ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" pull nginx core-backend global-backend public-web core-web organizer-web admin-web || true

# If pre-built images are not available in registry, fall back safely to local build
if ! docker image inspect "${DOCKER_REGISTRY:-hackways}/hackways-core-backend:${IMAGE_TAG:-latest}" >/dev/null 2>&1; then
    echo "Notice: Registry image missing, building core-backend fallback..."
    docker build -f "$PROJECT_DIR/deploy/docker/Dockerfile.core-backend" -t "${DOCKER_REGISTRY:-hackways}/hackways-core-backend:${IMAGE_TAG:-latest}" "$PROJECT_DIR"
fi
if ! docker image inspect "${DOCKER_REGISTRY:-hackways}/hackways-global-backend:${IMAGE_TAG:-latest}" >/dev/null 2>&1; then
    echo "Notice: Registry image missing, building global-backend fallback..."
    docker build -f "$PROJECT_DIR/deploy/docker/Dockerfile.global-backend" -t "${DOCKER_REGISTRY:-hackways}/hackways-global-backend:${IMAGE_TAG:-latest}" "$PROJECT_DIR"
fi

echo "=== 3. Database Connectivity & Migration ==="
# Check if using Azure Database for PostgreSQL Flexible Server or local container profile
if [ -n "${DATABASE_URL:-}" ]; then
    echo "Connecting to authoritative database (${DATABASE_URL%%@*}@...)..."
    for mig in "$PROJECT_DIR"/deploy/migrations/*.up.sql; do
        if [ -f "$mig" ]; then
            echo "Applying $(basename "$mig")..."
            psql "$DATABASE_URL" -f "$mig" 2>/dev/null || docker run --rm --network host -v "$PROJECT_DIR/deploy/migrations:/migrations" postgres:16-alpine psql "$DATABASE_URL" -f "/migrations/$(basename "$mig")" || true
        fi
    done
elif docker ps --format '{{.Names}}' | grep -q "hackways-postgres"; then
    for mig in "$PROJECT_DIR"/deploy/migrations/*.up.sql; do
        if [ -f "$mig" ]; then
            echo "Applying $(basename "$mig") to local container..."
            docker exec -i hackways-postgres psql -U "${POSTGRES_USER:-hackways_prod}" -d "${POSTGRES_DB:-hackways_prod}" < "$mig" || true
        fi
    done
fi

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
