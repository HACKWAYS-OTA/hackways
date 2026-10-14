#!/usr/bin/env bash
# ==============================================================================
# Hackways Azure VM Provisioning Script
# Target: Ubuntu 24.04 LTS (Standard_B2als_v2 / Standard_B2s / Standard_D2s_v5)
# ==============================================================================
set -euo pipefail

echo "=========================================="
echo " 1. Installing Official Docker Engine     "
echo "=========================================="
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg

sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu noble stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

echo "=========================================="
echo " 2. Configuring User Permissions & Group  "
echo "=========================================="
sudo usermod -aG docker azureuser

echo "=========================================="
echo " 3. Configuring Daemon Log Rotation Cap   "
echo "=========================================="
sudo tee /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

echo "=========================================="
echo " 4. Restarting Docker Daemon              "
echo "=========================================="
sudo systemctl daemon-reload
sudo systemctl restart docker
sudo systemctl enable docker

echo "=========================================="
echo " 5. Kernel Tuning & Swap Configuration    "
echo "=========================================="
echo -e "vm.swappiness = 10\nvm.vfs_cache_pressure = 50" | sudo tee /etc/sysctl.d/99-hackways.conf
sudo sysctl -p /etc/sysctl.d/99-hackways.conf

echo "=========================================="
echo " 6. Creating App Directories              "
echo "=========================================="
sudo mkdir -p /opt/hackways /var/log/hackways /etc/letsencrypt
sudo chown -R azureuser:azureuser /opt/hackways /var/log/hackways

echo "=========================================="
echo " Docker and Host Setup Complete!          "
echo "=========================================="
docker --version
docker compose version
