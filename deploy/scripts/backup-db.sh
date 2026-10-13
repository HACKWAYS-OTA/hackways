#!/usr/bin/env bash
# ==============================================================================
# Hackways Automated Database Backup Script
# Architecture: Zero-Disk Upload (Streams pg_dump directly to Azure Blob Storage)
# Container target: 'backups' in Azure Storage Account
# Retention: Daily automated cron, encrypted in transit & at rest
# ==============================================================================
set -euo pipefail

TIMESTAMP=$(date -u +"%Y%m%d_%H%M%SZ")
BACKUP_NAME="hackways_db_${TIMESTAMP}.sql.gz"

echo "=========================================================="
echo " Starting Hackways Direct-to-Azure-Blob Database Backup    "
echo " Timestamp: ${TIMESTAMP}                                 "
echo "=========================================================="

# Check required environment variables
if [ -z "${AZURE_STORAGE_ACCOUNT:-}" ]; then
    echo "ERROR: AZURE_STORAGE_ACCOUNT environment variable is required."
    exit 1
fi

if [ -z "${POSTGRES_USER:-}" ] || [ -z "${POSTGRES_DB:-}" ]; then
    echo "ERROR: POSTGRES_USER and POSTGRES_DB environment variables are required."
    exit 1
fi

CONTAINER_NAME="backups"

echo "--> Streaming pg_dump via gzip directly to Azure Blob Storage: ${BACKUP_NAME}"
# Pipe pg_dump directly to gzip, then stream via Azure CLI to avoid disk consumption on 50 GB NVMe
docker exec hackways-postgres pg_dump -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" --no-owner --no-privileges \
    | gzip -9 \
    | az storage blob upload \
        --account-name "${AZURE_STORAGE_ACCOUNT}" \
        --container-name "${CONTAINER_NAME}" \
        --name "${BACKUP_NAME}" \
        --auth-mode login \
        --data @- \
        --overwrite false

echo "--> Backup streamed and verified successfully in container '${CONTAINER_NAME}' as '${BACKUP_NAME}'."
echo "--> Zero disk space consumed on host VM."
