# Phase 9 Verification Protocol: Dual Go Backend Engines

## 1. Automated Verification Commands

```powershell
# 1. Verify Core Backend
Push-Location services/core-backend
gofmt -l .
go vet ./...
go test ./... -v -race -count=1
go build -o bin/core_backend ./cmd/server
Remove-Item -Force bin/core_backend
Pop-Location

# 2. Verify Global Backend
Push-Location services/global-backend
gofmt -l .
go vet ./...
go test ./... -v -race -count=1
go build -o bin/global_backend ./cmd/server
Remove-Item -Force bin/global_backend
Pop-Location
```

## 2. Acceptance Criteria
1. `services/core-backend`:
   - 0 unformatted files via `gofmt`.
   - 0 vet warnings via `go vet`.
   - Concurrency race tests against PostgreSQL pass 100% with race detector enabled.
   - Max connection pool capped at 28.
2. `services/global-backend`:
   - 0 unformatted files via `gofmt`.
   - 0 vet warnings via `go vet`.
   - Auth password tests, session tests, outbox worker tests, and Azure SAS tests pass 100% with race detector enabled.
   - Max connection pool capped at 22.
