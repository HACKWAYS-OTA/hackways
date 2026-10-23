ALTER TABLE events ADD COLUMN management_closed BOOLEAN NOT NULL DEFAULT FALSE;
-- Only recorded administrator publication closures are carried forward.
UPDATE events e SET management_closed=TRUE
WHERE e.status IN ('unpublished','cancelled') AND
 (SELECT a.action FROM audit_log a WHERE a.event_id=e.id AND a.action IN
 ('event.published','event.unpublished','event.cancelled','admin.event.published','admin.event.unpublished','admin.event.cancelled')
 ORDER BY a.id DESC LIMIT 1) IN ('admin.event.unpublished','admin.event.cancelled');
