#!/usr/bin/env bash
# ==============================================================================
# Hackways Disaster Recovery Restore Script
# Purpose: Streams compressed backup from Azure Blob Storage directly into PostgreSQL
# Safety: Requires explicit confirmation before overwriting active database
# ==============================================================================
set -euo pipefail

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <backup_blob_name>"
    echo "Example: $0 hackways_db_20261013_120000Z.sql.gz"
    exit 1
fi

BACKUP_BLOB="$1"

echo "=========================================================="
echo " WARNING: INITIATING DATABASE DISASTER RECOVERY RESTORE   "
echo " Target Backup: ${BACKUP_BLOB}                             "
echo "=========================================================="

read -p "Are you ABSOLUTELY sure you want to restore and overwrite the active database? (type 'RESTORE-CONFIRM'): " CONFIRMATION
if [ "${CONFIRMATION}" != "RESTORE-CONFIRM" ]; then
    echo "Restore aborted by user."
    exit 0
fi

if [ -z "${AZURE_STORAGE_ACCOUNT:-}" ]; then
    echo "ERROR: AZURE_STORAGE_ACCOUNT environment variable is required."
    exit 1
fi

CONTAINER_NAME="backups"

echo "--> Streaming compressed backup from Azure Blob Storage into PostgreSQL..."
az storage blob download \
    --account-name "${AZURE_STORAGE_ACCOUNT}" \
    --container-name "${CONTAINER_NAME}" \
    --name "${BACKUP_BLOB}" \
    --auth-mode login \
    --file - \
    | gunzip \
    | docker exec -i hackways-postgres psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}"

echo "--> Database restored successfully from ${BACKUP_BLOB}."
