# Deployment & Operations Guide — Hackways

---

## 1. Production Topology

```
                  ┌────────────────────────┐
                  │ Cloudflare / CDN Proxy  │
                  └───────────┬────────────┘
                              │ HTTPS
                  ┌───────────▼────────────┐
                  │ Reverse Proxy / Ingress│
                  └─────┬────────────┬─────┘
                        │            │
            port 3000   │            │ port 8080
      ┌─────────────────▼──┐      ┌──▼────────────────┐
      │ Next.js App Cluster│      │ Go API Binary Svc │
      └────────────────────┘      └─────────┬─────────┘
                                            │
                                            ▼
                                  ┌───────────────────┐
                                  │ PostgreSQL 16 RDS │
                                  └───────────────────┘
```

---

## 2. Environment Matrix

| Variable | Development | Staging | Production |
|---|---|---|---|
| `APP_ENV` | `development` | `staging` | `production` |
| `SESSION_COOKIE_SECURE` | `false` | `true` | `true` |
| `STORAGE_BACKEND` | `local` | `s3` | `s3` |
| `DB_SSLMODE` | `disable` | `require` | `verify-full` |

---

## 3. Migration Rollout Strategy

1. Migrations are executed prior to traffic switch.
2. Every migration file (`00000X_*.up.sql`) must have a corresponding, tested rollback script (`00000X_*.down.sql`).
3. Schema modifications must be backwards-compatible with running binary instances (expand-contract pattern).

---

## 4. Graceful Shutdown & Drain

The Go API server catches `SIGTERM` and `SIGINT`:
1. Closes HTTP listener to reject new connections.
2. Waits up to 30 seconds to drain active in-flight HTTP requests.
3. Closes database pool (`pgxpool.Close()`).
4. Exits with status 0.
