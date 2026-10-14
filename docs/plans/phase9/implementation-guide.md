# Phase 9 Implementation Guide: Dual Go Backend Engines

## 1. Directory Structure

```
services/
├── core-backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── events/
│   │   ├── rsvps/
│   │   └── platform/
│   ├── db/
│   ├── go.mod
│   └── sqlc.yaml
└── global-backend/
    ├── cmd/server/main.go
    ├── internal/
    │   ├── auth/
    │   ├── storage/
    │   ├── notifications/
    │   └── platform/
    ├── db/
    ├── go.mod
    └── sqlc.yaml
```

---

## 2. Decoupling Rules & Invariants
1. `core-backend` must never import `auth` session password hashing or Azure SAS generation directly. It consumes authentication context passed via request headers or internal verification.
2. `global-backend` must never execute event reservation row locks.
3. Both services use compile-time verified `sqlc` queries mapped to their respective domain responsibilities.
4. Database connection limits must be strictly hard-coded in defaults (`core-backend`: 28, `global-backend`: 22).
