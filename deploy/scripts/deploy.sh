#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Hackways Production Deployment Automation
# Target: Azure VM (Standard B2als v2 - 2 vCPU, 4 GB RAM, 50 GB NVMe)
# ==============================================================================

PROJECT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
COMPOSE_FILE="$PROJECT_DIR/deploy/docker/docker-compose.prod.yml"
ENV_FILE="${HACKWAYS_ENV_FILE:-$PROJECT_DIR/.env}"

echo "=== 1. Validating Production Environment ==="
if [ ! -f "$ENV_FILE" ]; then
    echo "ERROR: Environment file $ENV_FILE not found!"
    exit 1
fi

: "${DEPLOY_IMAGE_TAG:?Set DEPLOY_IMAGE_TAG to the reviewed full Git commit SHA}"
if [[ ! "$DEPLOY_IMAGE_TAG" =~ ^[0-9a-f]{40}$ ]]; then
    echo "ERROR: DEPLOY_IMAGE_TAG must be a full lowercase Git commit SHA."
    exit 1
fi
export IMAGE_TAG="$DEPLOY_IMAGE_TAG"

command -v python3 >/dev/null 2>&1 || {
    echo "ERROR: Python 3 is required for protected migration configuration."
    exit 1
}
command -v migrate >/dev/null 2>&1 || {
    echo "ERROR: golang-migrate CLI is required on the deployment host."
    exit 1
}

# Compose owns dotenv parsing. Never execute the environment file as shell code.
# Validate without emitting the resolved configuration, which contains secrets.
compose=(docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE")
"${compose[@]}" config --quiet

echo "=== 2. Pulling CI-built images for reviewed commit $IMAGE_TAG ==="
"${compose[@]}" pull

echo "=== 3. Applying pending database migrations ==="
"${compose[@]}" config --format json | python3 "$PROJECT_DIR/deploy/scripts/run-migrations.py" "$PROJECT_DIR/deploy/migrations"

echo "=== 4. Launching production containers and waiting for health ==="
"${compose[@]}" up -d --wait --wait-timeout 120

# Frontend services do not yet define Docker healthchecks. Probe them explicitly.
for target in public-web:3000 core-web:3001 organizer-web:3002 admin-web:3003; do
    service="${target%:*}"
    port="${target##*:}"
    "${compose[@]}" exec -T "$service" wget -q -T 10 -O /dev/null "http://127.0.0.1:$port/"
done

echo "Reloading Nginx ingress configuration..."
docker exec hackways-nginx nginx -t
docker exec hackways-nginx nginx -s reload

echo "=== 5. Live Container Health and Resource Inspection ==="
"${compose[@]}" ps

echo "=== Hackways Production Deployment Complete ==="
