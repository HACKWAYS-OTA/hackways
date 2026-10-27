# Administrator event Bin

Core Go owns the recoverable event workflow. The existing dedicated admin
session, database grant and daily verification protect every route. Mutations
also require the approved Origin, random idempotency key and audited reason.

- `GET /api/v1/admin/core/bin?q=&limit=20&offset=0` lists soft-deleted events,
  including older owner-deleted drafts, newest deletion first. Limit is1-100.
- `POST /api/v1/admin/core/events/{reference}/bin` accepts `{revision}`.
- `POST /api/v1/admin/core/bin/events/{reference}/restore` accepts `{revision}`.

Every write locks the event row, including deleted rows, with the existing
two-second statement timeout, rechecks admin authority, checks the revision and
stores its receipt/audit in the same transaction. Exact retries replay after
deletion/restoration; changing the request under that key conflicts. No schema
change and no permanent purge are included.

Published events that have not ended and other active events with bookings must
first use the existing cancellation flow, which owns guest notification and
inventory handling. Inactive drafts/unpublished events and ended historical
events can enter Bin. The operation retains registrations, tickets, access
assignments, curation, image references, audit and the event row. Public and
workspace catalogues continue using their established deleted-event exclusion.

Restoration clears deletion, increments revision and changes published status to
unpublished. Cancelled status, bookings, ticket states and inventory stay as
persisted. Recovery never republishes or reactivates a cancelled registration.

The console exposes Bin search/pagination and per-record recovery with a reason
and shared confirmation control. Event details expose Move to Bin. Confirmed
mutations navigate to Bin or the recovered event; errors preserve entered data.
Initial loading reuses the Hackways wordmark animation.
