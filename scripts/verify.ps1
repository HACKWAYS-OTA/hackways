# ==============================================================================
# Hackways Repository PowerShell Verification Suite
# ==============================================================================
$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " 1. Go Backend: Format, Vet, Tests, Build " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Push-Location services/api

Write-Host "--> Checking Go formatting..."
$unformatted = gofmt -l .
if ($unformatted) {
    Write-Error "ERROR: Unformatted Go files: $unformatted"
    exit 1
}
Write-Host "Go formatting clean." -ForegroundColor Green

Write-Host "--> Running go vet..."
go vet ./...
Write-Host "Go vet clean." -ForegroundColor Green

Write-Host "--> Running Go test suite..."
go test ./... -v -count=1
Write-Host "Go tests passed." -ForegroundColor Green

Write-Host "--> Compiling Go API binary..."
go build -o bin/api_test ./cmd/api
Remove-Item -Force bin/api_test
Write-Host "Backend compiled cleanly." -ForegroundColor Green

Pop-Location

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host " 2. Frontend: Syntax, Types & ESLint      " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$apps = @("core-web", "public-web", "organizer-web", "admin-web")
foreach ($app in $apps) {
    Write-Host "--> Checking apps/$app..."
    Push-Location "apps/$app"
    npm run type-check
    npm run lint
    Pop-Location
    Write-Host "apps/$app verified." -ForegroundColor Green
}

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host " 3. Repository Hygiene & Secret Audit     " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

if (Test-Path .env) {
    Write-Warning "WARNING: Local .env exists (must remain gitignored)."
}

Write-Host "`nALL CHECKS PASSED SUCCESSFULLY." -ForegroundColor Green
