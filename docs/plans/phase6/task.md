# Phase 6 Task Register: Frontend Foundations

## Tasks Checklist

| Task ID | Title | Prerequisites | Status | Owner | Blockers / References |
|:---:|:---|:---:|:---:|:---:|:---|
| `P6-FE-001` | Refactor `apps/web` into `apps/core-web` with standalone config | Phase 5 Exit Gate | `COMPLETED` | Frontend Lead | AD-001 |
| `P6-FE-002` | Scaffold `apps/public-web` for SSG/ISR marketing & catalogs | `P6-FE-001` | `COMPLETED` | Frontend Lead | Port 3000 |
| `P6-FE-003` | Scaffold `apps/organizer-web` for event studio & attendee roster | `P6-FE-001` | `COMPLETED` | Frontend Lead | Port 3002 |
| `P6-FE-004` | Scaffold `apps/admin-web` for system audits & telemetry | `P6-FE-001` | `COMPLETED` | Frontend Lead | Port 3003 |
| `P6-FE-005` | Purge `SAMPLE_EVENTS` mock data and replace with server fetch | `P6-FE-001` | `COMPLETED` | Frontend Lead | GAP-006, `AGENTS.md` 1.1 |
| `P6-FE-006` | Connect `@hackways/api-client` to all frontend applications | `P6-FE-002`, `003`, `004` | `COMPLETED` | Tooling Lead | Shared SDK |
| `P6-FE-007` | Implement accessible UI primitives without `lucide-react` | `P6-FE-001` | `COMPLETED` | UI Engineer | Inline SVGs |
| `P6-FE-008` | Implement loading skeletons and error boundary retry states | `P6-FE-005` | `COMPLETED` | UI Engineer | `AGENTS.md` 1.4 |
| `P6-FE-009` | Execute Phase 6 exit gate verification across all 4 frontends | All P6 Tasks | `COMPLETED` | Tech Lead | Phase Gate 6 |

---

## Detailed Task Specifications

### `P6-FE-001`: Refactor `apps/web` into `apps/core-web`
- **Description:** Rename/relocate `apps/web` to `apps/core-web` with port 3001, targeting `app.hackways.in`. Configure `output: 'standalone'` and memory limit `--max-old-space-size=190`.
- **Acceptance Criteria:** `core-web` runs independently on port 3001.
- **Status:** `NOT_STARTED`.

### `P6-FE-002`: Scaffold `apps/public-web`
- **Description:** Establish Next.js 14 SSG/ISR project on port 3000 targeting `hackways.in` for public SEO landing pages. Memory cap: 140 MB (`--max-old-space-size=100`).
- **Acceptance Criteria:** `public-web` builds standalone and serves landing shell.
- **Status:** `NOT_STARTED`.

### `P6-FE-003`: Scaffold `apps/organizer-web`
- **Description:** Establish Next.js 14 App Router project on port 3002 targeting `organizer.hackways.in` for event studio and roster management. Memory cap: 230 MB (`--max-old-space-size=180`).
- **Acceptance Criteria:** `organizer-web` serves organizer studio shell.
- **Status:** `NOT_STARTED`.

### `P6-FE-004`: Scaffold `apps/admin-web`
- **Description:** Establish Next.js 14 App Router project on port 3003 targeting `admin.hackways.in` with IP/VPN restriction checks. Memory cap: 170 MB (`--max-old-space-size=130`).
- **Acceptance Criteria:** `admin-web` serves admin shell.
- **Status:** `NOT_STARTED`.

### `P6-FE-005`: Purge mock data
- **Description:** Delete `SAMPLE_EVENTS` in `app/page.tsx` and replace `handleRegister` browser `alert()` with real form submission calling `POST /api/v1/rsvps`.
- **Acceptance Criteria:** Zero mock data in code; real API response rendered.
- **Status:** `NOT_STARTED`.

### `P6-FE-006`: Connect shared API client
- **Description:** Link `@hackways/api-client` package to all 4 frontend applications.
- **Acceptance Criteria:** Shared types and HTTP client methods imported cleanly without type errors.
- **Status:** `NOT_STARTED`.

### `P6-FE-007`: Accessible UI primitives
- **Description:** Create accessible `Button`, `Badge`, `Input`, `Dialog`, and `EventCard` primitives using Tailwind CSS and native inline SVG icons.
- **Acceptance Criteria:** Accessible form controls; 0 imports of `lucide-react`.
- **Status:** `NOT_STARTED`.

### `P6-FE-008`: Loading skeletons & error recovery
- **Description:** Implement React Suspense loading skeletons for event cards and error boundaries with retry buttons.
- **Acceptance Criteria:** Zero layout shifts on first render; clear error recovery UI.
- **Status:** `NOT_STARTED`.

### `P6-FE-009`: Phase 6 exit gate verification
- **Description:** Confirm all criteria in `docs/plans/phase6/verification.md`.
- **Acceptance Criteria:** Exit gate approved; Phase 6 marked `VERIFIED`.
- **Status:** `NOT_STARTED`.
