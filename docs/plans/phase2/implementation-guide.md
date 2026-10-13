# Phase 2 Implementation Guide: Toolchains & Dependency Hygiene

## 1. Prerequisites
- Phase 1 verified.
- Node.js 20+ and npm installed.
- Go 1.23+ installed.

---

## 2. Step-by-Step Implementation

### Step 1: Purge `lucide-react`
1. Open `apps/web/package.json`.
2. Delete the line:
   ```json
   "lucide-react": "^0.460.0",
   ```
3. Remove `lucide-react` entries from `apps/web/package-lock.json`.
4. Run `npm install` in `apps/web` to synchronize lockfiles cleanly.
5. Verify with:
   ```bash
   npm ls lucide-react
   ```
   (Must output `(empty)`).

### Step 2: Configure Standalone Output in Next.js
1. Open `apps/web/next.config.js`.
2. Update config to include `output: 'standalone'`:
   ```javascript
   /** @type {import('next').NextConfig} */
   const nextConfig = {
     output: 'standalone',
     reactStrictMode: true,
     poweredByHeader: false,
   };

   module.exports = nextConfig;
   ```

### Step 3: Add `pgx/v5` to Go Backend
1. In `services/api`, run:
   ```bash
   go get github.com/jackc/pgx/v5
   go mod tidy
   ```
2. Verify that `services/api/go.mod` lists `github.com/jackc/pgx/v5`.
3. Verify compilation:
   ```bash
   go build -o /dev/null ./cmd/api
   ```

### Step 4: Run Static Checks
1. Backend: `go vet ./...` in `services/api`.
2. Frontend: `npm run type-check` in `apps/web`.

---

## 3. What Must NOT Be Changed in Phase 2
- Do not add alternative icon libraries until approved (DEC-005).
- Do not introduce ORMs (GORM, Ent) into Go modules.
- Do not run `next build` on the host VM if memory is low.
