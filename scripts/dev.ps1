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

Write-Host "==> Starting backend Go API..." -ForegroundColor Green
Start-Process -FilePath "go" -ArgumentList "run", "./cmd/api" -WorkingDirectory "services\api"

Write-Host "==> Starting frontend Next.js dev server..." -ForegroundColor Green
Start-Process -FilePath "npm" -ArgumentList "run", "dev" -WorkingDirectory "apps\web"

Write-Host "Services started." -ForegroundColor Green
