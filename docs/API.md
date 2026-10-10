# REST API Specification — Hackways

**Base Path:** `/api/v1`  
**Data Format:** JSON  
**Authentication:** Cookie `hackways_session` (HttpOnly, Secure in prod)  

---

## 1. Global Conventions

### Standard Response Envelope
All success and error responses follow deterministic formats:

**Success (Collections):**
```json
{
  "data": [...],
  "pagination": {
    "cursor": "eyJpZCI6Ii4uLiJ9",
    "has_more": true,
    "total": 42
  }
}
```

**Error Envelope:**
```json
{
  "error": {
    "code": "EVENT_AT_CAPACITY",
    "message": "This event has reached its maximum registration capacity.",
    "details": {}
  }
}
```

### Standard Error Codes
| Code | HTTP Status | Description |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Invalid request syntax or schema |
| `UNAUTHORIZED` | 401 | Missing or invalid session cookie |
| `FORBIDDEN` | 403 | Authenticated user lacks permission |
| `NOT_FOUND` | 404 | Resource does not exist |
| `CONFLICT` | 409 | Duplicate entity or state conflict |
| `EVENT_AT_CAPACITY`| 409 | Event capacity reached |
| `UNPROCESSABLE` | 422 | Semantic business rule violation |
| `RATE_LIMITED` | 429 | Rate limit exceeded |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## 2. Endpoints Summary

### Health
- `GET /health` — Service readiness and PostgreSQL connection check.

### Authentication (`/api/v1/auth`)
- `POST /api/v1/auth/signup` — Create user account & establish session cookie.
- `POST /api/v1/auth/login` — Authenticate credentials & issue session cookie.
- `DELETE /api/v1/auth/logout` — Invalidate session and clear cookie.
- `GET /api/v1/auth/me` — Retrieve authenticated user profile and organizer status.

### Events (`/api/v1/events`)
- `GET /api/v1/events` — Public event listing with category, search query, format, and date filters.
- `GET /api/v1/events/:id` — Public event detail.
- `POST /api/v1/events` — Create new event draft (Organizer only).
- `PATCH /api/v1/events/:id` — Update event attributes (Owner only).
- `POST /api/v1/events/:id/publish` — Transition event to published state.
- `POST /api/v1/events/:id/unpublish` — Revert event to unpublished state.
- `POST /api/v1/events/:id/cancel` — Cancel event & queue attendee notifications.
- `POST /api/v1/events/:id/banner` — Upload event banner image (`multipart/form-data`).
- `GET /api/v1/events/:id/registrations` — Private attendee list (Owner only).

### Registrations & RSVPs (`/api/v1/registrations`)
- `POST /api/v1/registrations` — Register for an event (Guest or authenticated user).
- `GET /api/v1/registrations/:id` — Registration details (Registrant or event owner).
- `DELETE /api/v1/registrations/:id` — Cancel registration.

### Organizers (`/api/v1/organizers`)
- `POST /api/v1/organizers` — Create organizer profile for user.
- `GET /api/v1/organizers/:slug` — Public organizer profile.
- `PATCH /api/v1/organizers/me` — Update organizer details.
- `GET /api/v1/organizers/me/events` — Organizer's events across all statuses.
