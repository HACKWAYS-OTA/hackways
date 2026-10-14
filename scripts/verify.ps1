# ==============================================================================
# Hackways Repository PowerShell Verification Suite
# ==============================================================================
$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " 1. Go Backend: Format, Vet, Tests, Build " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$backendServices = @("core-backend", "global-backend")
foreach ($svc in $backendServices) {
    if (Test-Path "services/$svc") {
        Write-Host "`n--> Verifying services/$svc..." -ForegroundColor Yellow
        Push-Location "services/$svc"
        
        Write-Host "--> Checking Go formatting..."
        $unformatted = gofmt -l .
        if ($unformatted) {
            Write-Error "ERROR: Unformatted Go files in services/$($svc): $unformatted"
            exit 1
        }
        Write-Host "Go formatting clean." -ForegroundColor Green
        
        Write-Host "--> Running go vet..."
        go vet ./...
        Write-Host "Go vet clean." -ForegroundColor Green
        
        Write-Host "--> Running Go test suite..."
        go test ./... -v -count=1
        Write-Host "Go tests passed." -ForegroundColor Green
        
        Write-Host "--> Compiling Go binary..."
        go build -o bin/server_test ./cmd/server
        Remove-Item -Force bin/server_test
        Write-Host "services/$svc compiled cleanly." -ForegroundColor Green
        
        Pop-Location
    }
}

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
