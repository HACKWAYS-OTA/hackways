#!/usr/bin/env bash
# ==============================================================================
# Hackways Host Swap & Kernel Tuning Provisioning Script
# Target: Azure VM (2 vCPU, 4 GB RAM, 50 GB NVMe)
# Memory Governance: 3.0 GB NVMe emergency swapfile, vm.swappiness = 10
# ==============================================================================
set -euo pipefail

echo "=========================================================="
echo " Provisioning 3.0 GB NVMe Swapfile & Tuning Host Kernel    "
echo "=========================================================="

SWAPFILE="/swapfile"

if [ -f "${SWAPFILE}" ]; then
    echo "Swapfile already exists at ${SWAPFILE}. Skipping allocation."
else
    echo "--> Allocating 3.0 GB swapfile on NVMe disk..."
    fallocate -l 3G "${SWAPFILE}" || dd if=/dev/zero of="${SWAPFILE}" bs=1M count=3072
    chmod 600 "${SWAPFILE}"
    mkswap "${SWAPFILE}"
    swapon "${SWAPFILE}"
    echo "${SWAPFILE} none swap sw 0 0" >> /etc/fstab
    echo "Swapfile successfully created and activated."
fi

echo "--> Applying kernel sysctl memory governance parameters..."
cat << 'EOF' > /etc/sysctl.d/99-hackways-memory.conf
# Keep swappiness low so swap is used only as an emergency safety net
vm.swappiness = 10

# Maintain dirty pages below 10% to prevent I/O write freezes
vm.dirty_ratio = 10
vm.dirty_background_ratio = 5

# Protect dentries and inodes in page cache
vm.vfs_cache_pressure = 50

# Prevent memory overcommit exhaustion
vm.overcommit_memory = 0
EOF

sysctl -p /etc/sysctl.d/99-hackways-memory.conf
echo "--> Kernel parameters applied successfully."

echo "--> Current active swap configuration:"
swapon --show
