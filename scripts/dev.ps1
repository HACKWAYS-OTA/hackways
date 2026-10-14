# ==============================================================================
# Hackways Local Development Launcher (PowerShell)
# ==============================================================================

Write-Host "==> Checking prerequisites..." -ForegroundColor Cyan

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go is required but not installed on PATH."
    exit 1
}

if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Error "Node.js is required but not installed on PATH."
    exit 1
}

Write-Host "==> Starting Core Backend (port 8081)..." -ForegroundColor Green
Start-Process -FilePath "go" -ArgumentList "run", "./cmd/server" -WorkingDirectory "services\core-backend"

Write-Host "==> Starting Global Backend (port 8082)..." -ForegroundColor Green
Start-Process -FilePath "go" -ArgumentList "run", "./cmd/server" -WorkingDirectory "services\global-backend"

Write-Host "==> Starting Core Web frontend (port 3001)..." -ForegroundColor Green
Start-Process -FilePath "npm" -ArgumentList "run", "dev" -WorkingDirectory "apps\core-web"

Write-Host "Services started." -ForegroundColor Green
