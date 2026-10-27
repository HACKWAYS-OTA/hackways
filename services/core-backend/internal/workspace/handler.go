// Package workspace exposes owner-managed collaboration without platform powers.
package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool     *pgxpool.Pool
	q        *db.Queries
	sessions *auth.SessionManager
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	q := db.New(pool)
	return &Handler{pool, q, auth.NewSessionManager(q)}
}
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(auth.RequireAuth(h.sessions))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "private, no-store")
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/spaces", h.Spaces)
	r.Get("/events", h.Events)
	for _, kind := range []string{"space", "event"} {
		gate := auth.RequireSpacePermission(h.q, "access.manage")
		if kind == "event" {
			gate = auth.RequireEventPermission(h.q, "access.manage")
		}
		r.With(gate).Get("/"+kind+"s/{id}/access", h.Access)
		r.With(gate).Post("/"+kind+"s/{id}/access", h.Grant)
		r.With(gate).Delete("/"+kind+"s/{id}/access/{access}", h.Revoke)
	}
	return r
}
func pagination(w http.ResponseWriter, r *http.Request) (int32, int32, string, bool) {
	limit, offset := int64(30), int64(0)
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.ParseInt(raw, 10, 32)
	}
	if err != nil || limit < 1 || limit > 100 {
		platform.RespondError(w, 400, "INVALID_PAGE", "Choose 1 to 100 records", nil)
		return 0, 0, "", false
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.ParseInt(raw, 10, 32)
	}
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	if err != nil || offset < 0 || len([]rune(search)) > 200 {
		platform.RespondError(w, 400, "INVALID_PAGE", "Check the page and search", nil)
		return 0, 0, "", false
	}
	return int32(limit), int32(offset), search, true
}
func (h *Handler) Spaces(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	limit, offset, search, ok := pagination(w, r)
	if !ok {
		return
	}
	user, _ := auth.UserFromContext(r.Context())
	rows, err := h.q.ListWorkspaceSpaces(ctx, db.ListWorkspaceSpacesParams{ActorID: user.UserID, Q: search, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.ListWorkspaceSpacesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data   []db.ListWorkspaceSpacesRow `json:"data"`
		Limit  int32                       `json:"limit"`
		Offset int32                       `json:"offset"`
	}{rows, limit, offset})
}
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	limit, offset, search, ok := pagination(w, r)
	if !ok {
		return
	}
	user, _ := auth.UserFromContext(r.Context())
	view := r.URL.Query().Get("view")
	if view == "" {
		view = "all"
	}
	if view != "all" && view != "drafts" && view != "upcoming" && view != "past" && view != "cancelled" {
		platform.RespondError(w, 400, "INVALID_VIEW", "Choose a valid event view", nil)
		return
	}
	rows, err := h.q.ListWorkspaceEvents(ctx, db.ListWorkspaceEventsParams{ActorID: user.UserID, Q: search, View: view, SpaceSlug: r.URL.Query().Get("space"), PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.ListWorkspaceEventsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data   []db.ListWorkspaceEventsRow `json:"data"`
		Limit  int32                       `json:"limit"`
		Offset int32                       `json:"offset"`
	}{rows, limit, offset})
}

type member struct {
	ID        pgtype.UUID        `json:"id"`
	UserID    pgtype.UUID        `json:"user_id"`
	Name      string             `json:"display_name"`
	Email     string             `json:"email"`
	Role      string             `json:"role"`
	Revision  int64              `json:"revision"`
	RevokedAt pgtype.Timestamptz `json:"revoked_at"`
}

func (h *Handler) Access(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	limit, offset, _, ok := pagination(w, r)
	if !ok {
		return
	}
	access, _ := auth.WorkspaceFromContext(r.Context())
	members, inherited := []member{}, []member{}
	if access.Kind == "space" {
		rows, err := h.q.ListSpaceAccess(ctx, db.ListSpaceAccessParams{SpaceID: access.ResourceID, Limit: limit, Offset: offset})
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		for _, row := range rows {
			members = append(members, member{row.ID, row.UserID, row.DisplayName, row.Email, row.Role, row.Revision, row.RevokedAt})
		}
	} else {
		rows, err := h.q.ListEventAccess(ctx, db.ListEventAccessParams{EventID: access.ResourceID, Limit: limit, Offset: offset})
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		for _, row := range rows {
			members = append(members, member{row.ID, row.UserID, row.DisplayName, row.Email, row.Role, row.Revision, row.RevokedAt})
		}
		inheritedRows, err := h.q.ListInheritedEventAccess(ctx, db.ListInheritedEventAccessParams{ID: access.ResourceID, Limit: limit, Offset: offset})
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		for _, row := range inheritedRows {
			inherited = append(inherited, member{row.ID, row.UserID, row.DisplayName, row.Email, row.Role, row.Revision, row.RevokedAt})
		}
	}
	platform.RespondJSON(w, 200, struct {
		Data      []member `json:"data"`
		Inherited []member `json:"inherited"`
		Limit     int32    `json:"limit"`
		Offset    int32    `json:"offset"`
	}{members, inherited, limit, offset})
}
func (h *Handler) begin(r *http.Request) (pgx.Tx, *db.Queries, error) {
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		return nil, nil, err
	}
	q := h.q.WithTx(tx)
	if err = q.SetWorkflowTimeout(r.Context()); err != nil {
		tx.Rollback(r.Context())
		return nil, nil, err
	}
	access, _ := auth.WorkspaceFromContext(r.Context())
	owner, _ := auth.OrganizerFromContext(r.Context())
	var actual pgtype.UUID
	if access.Kind == "event" {
		row, e := q.LockWorkflowEvent(r.Context(), access.ResourceID)
		err = e
		actual = row.OrganizerID
	} else {
		row, e := q.LockSpace(r.Context(), access.ResourceID)
		err = e
		actual = row.OrganizerID
	}
	if err == nil && actual != owner.ID {
		err = pgx.ErrNoRows
	}
	if err == nil {
		err = auth.RecheckWorkspace(r.Context(), q, access.Kind, access.ResourceID)
	}
	if err != nil {
		tx.Rollback(r.Context())
		return nil, nil, err
	}
	return tx, q, nil
}
func fail(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 409, "ACCESS_CHANGED", "Access changed. Reload the team before trying again", nil)
	} else if errors.Is(err, auth.ErrWorkspaceAccessDenied) {
		platform.RespondError(w, 403, "WORKSPACE_ACCESS_DENIED", "Your workspace access changed", nil)
	} else {
		platform.RespondDatabaseError(w, err)
	}
}

type grant struct {
	ID        pgtype.UUID        `json:"id"`
	UserID    pgtype.UUID        `json:"user_id"`
	Role      string             `json:"role"`
	Revision  int64              `json:"revision"`
	RevokedAt pgtype.Timestamptz `json:"revoked_at"`
}

func readGrant(ctx context.Context, q *db.Queries, access auth.WorkspaceAccess, user, id pgtype.UUID) (grant, error) {
	if access.Kind == "space" {
		var row db.SpaceAccess
		var err error
		if id.Valid {
			row, err = q.ReadSpaceAccessByID(ctx, db.ReadSpaceAccessByIDParams{SpaceID: access.ResourceID, ID: id})
		} else {
			row, err = q.ReadSpaceAccess(ctx, db.ReadSpaceAccessParams{SpaceID: access.ResourceID, UserID: user})
		}
		return grant{row.ID, row.UserID, row.Role, row.Revision, row.RevokedAt}, err
	}
	var row db.EventAccess
	var err error
	if id.Valid {
		row, err = q.ReadEventAccessByID(ctx, db.ReadEventAccessByIDParams{EventID: access.ResourceID, ID: id})
	} else {
		row, err = q.ReadEventAccess(ctx, db.ReadEventAccessParams{EventID: access.ResourceID, UserID: user})
	}
	return grant{row.ID, row.UserID, row.Role, row.Revision, row.RevokedAt}, err
}
func (h *Handler) Grant(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Role     string `json:"role"`
		Revision int64  `json:"revision"`
	}
	if platform.DecodeJSON(r, &body) != nil {
		platform.RespondError(w, 400, "INVALID_PAYLOAD", "Provide an account email and role", nil)
		return
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	email, err := mail.ParseAddress(body.Email)
	if err != nil || email.Address != body.Email || len(body.Email) > 254 || body.Revision < 0 || (body.Role != "manager" && body.Role != "editor" && body.Role != "checkin") {
		platform.RespondError(w, 422, "INVALID_ACCESS", "Choose an account email and a supported role", nil)
		return
	}
	tx, q, err := h.begin(r)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	account, err := q.WorkspaceAccountByEmail(r.Context(), body.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 422, "ACCOUNT_NOT_AVAILABLE", "This person needs a verified Hackways account before access can be assigned", nil)
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	owner, _ := auth.OrganizerFromContext(r.Context())
	access, _ := auth.WorkspaceFromContext(r.Context())
	if account.ID == owner.UserID {
		platform.RespondError(w, 409, "OWNER_ACCESS", "The owner already has full access", nil)
		return
	}
	current, err := readGrant(r.Context(), q, access, account.ID, pgtype.UUID{})
	missing := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !missing {
		fail(w, err)
		return
	}
	if !missing && current.Role == body.Role && !current.RevokedAt.Valid {
		platform.RespondJSON(w, 200, struct {
			Data grant `json:"data"`
		}{current})
		return
	}
	if missing && body.Revision != 0 || !missing && current.Revision != body.Revision {
		fail(w, pgx.ErrNoRows)
		return
	}
	if access.Kind == "space" {
		var row db.SpaceAccess
		if missing {
			row, err = q.CreateSpaceAccess(r.Context(), db.CreateSpaceAccessParams{SpaceID: access.ResourceID, UserID: account.ID, Role: body.Role, GrantedBy: actor.UserID})
		} else {
			row, err = q.UpdateSpaceAccess(r.Context(), db.UpdateSpaceAccessParams{SpaceID: access.ResourceID, UserID: account.ID, Role: body.Role, UpdatedBy: actor.UserID, Revision: body.Revision})
		}
		current = grant{row.ID, row.UserID, row.Role, row.Revision, row.RevokedAt}
	} else {
		var row db.EventAccess
		if missing {
			row, err = q.CreateEventAccess(r.Context(), db.CreateEventAccessParams{EventID: access.ResourceID, UserID: account.ID, Role: body.Role, GrantedBy: actor.UserID})
		} else {
			row, err = q.UpdateEventAccess(r.Context(), db.UpdateEventAccessParams{EventID: access.ResourceID, UserID: account.ID, Role: body.Role, UpdatedBy: actor.UserID, Revision: body.Revision})
		}
		current = grant{row.ID, row.UserID, row.Role, row.Revision, row.RevokedAt}
	}
	if err == nil {
		err = audit(r, q, access, current.UserID, current.Role, "access.granted")
	}
	if err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		fail(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data grant `json:"data"`
	}{current})
}
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Revision int64 `json:"revision"`
	}
	var id pgtype.UUID
	if platform.DecodeJSON(r, &body) != nil || body.Revision < 1 || id.Scan(chi.URLParam(r, "access")) != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ACCESS", "Provide the current access record", nil)
		return
	}
	tx, q, err := h.begin(r)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	access, _ := auth.WorkspaceFromContext(r.Context())
	actor, _ := auth.UserFromContext(r.Context())
	current, err := readGrant(r.Context(), q, access, pgtype.UUID{}, id)
	if err != nil {
		fail(w, err)
		return
	}
	if current.RevokedAt.Valid {
		platform.RespondJSON(w, 200, struct {
			Data grant `json:"data"`
		}{current})
		return
	}
	if current.Revision != body.Revision {
		fail(w, pgx.ErrNoRows)
		return
	}
	if access.Kind == "space" {
		row, e := q.RevokeSpaceAccess(r.Context(), db.RevokeSpaceAccessParams{SpaceID: access.ResourceID, ID: id, Revision: body.Revision, UpdatedBy: actor.UserID})
		err = e
		current = grant{row.ID, row.UserID, row.Role, row.Revision, row.RevokedAt}
	} else {
		row, e := q.RevokeEventAccess(r.Context(), db.RevokeEventAccessParams{EventID: access.ResourceID, ID: id, Revision: body.Revision, UpdatedBy: actor.UserID})
		err = e
		current = grant{row.ID, row.UserID, row.Role, row.Revision, row.RevokedAt}
	}
	if err == nil {
		err = audit(r, q, access, current.UserID, current.Role, "access.revoked")
	}
	if err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		fail(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data grant `json:"data"`
	}{current})
}
func audit(r *http.Request, q *db.Queries, access auth.WorkspaceAccess, target pgtype.UUID, role, action string) error {
	actor, _ := auth.UserFromContext(r.Context())
	event := pgtype.UUID{}
	if access.Kind == "event" {
		event = access.ResourceID
	}
	metadata, err := json.Marshal(struct {
		UserID pgtype.UUID `json:"target_user_id"`
		Role   string      `json:"role"`
	}{target, role})
	if err != nil {
		return err
	}
	return q.AppendWorkflowAudit(r.Context(), db.AppendWorkflowAuditParams{ActorUserID: actor.UserID, EventID: event, EntityID: access.ResourceID, Action: access.Kind + "." + action, RequestID: platform.GetRequestID(r.Context()), Metadata: metadata})
}
