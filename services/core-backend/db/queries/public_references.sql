-- name: GetAccountRegistrationByCode :one
SELECT * FROM rsvps WHERE public_code=$1 AND user_id=$2;

-- name: GetEventTicketByCode :one
SELECT t.id FROM tickets t JOIN rsvps r ON r.id=t.rsvp_id WHERE t.public_code=$1 AND r.event_id=$2;
