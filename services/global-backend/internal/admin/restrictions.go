package admin

import (
	"net/http"
	"strconv"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) RestrictedAccounts(w http.ResponseWriter, r *http.Request) {
	query, _, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListRestrictedAccounts(r.Context(), db.AdminListRestrictedAccountsParams{Q: query, PageLimit: limit, PageOffset: offset})
	if err != nil {
		detailError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListRestrictedAccountsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListRestrictedAccountsRow `json:"data"`
	}{rows})
}
func (h *Handler) RestrictedDevices(w http.ResponseWriter, r *http.Request) {
	query, _, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListRestrictedDevices(r.Context(), db.AdminListRestrictedDevicesParams{Q: query, PageLimit: limit, PageOffset: offset})
	if err != nil {
		detailError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListRestrictedDevicesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListRestrictedDevicesRow `json:"data"`
	}{rows})
}

// The caller holds the role/action advisory locks and has checked fresh admin
// authorization and receipt replay. Device locks always precede user row locks.
func (h *Handler) applyRestriction(w http.ResponseWriter, r *http.Request, q *db.Queries, action string, id, actor pgtype.UUID, body actionBody, reason string) (actionAudit, error, bool) {
	metadata := actionAudit{Reason: reason}
	deny := func(status int, code, message string) (actionAudit, error, bool) {
		platform.RespondError(w, status, code, message, nil)
		return metadata, nil, false
	}
	if action == "device.restriction" {
		if *body.Blocked {
			return deny(422, "INVALID_ACTION", "Block a recognized browser from its account record")
		}
		hash, err := q.AdminReadRestrictedDevice(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return metadata, nil, false
		}
		if err = q.LockRestrictionDevice(r.Context(), hash); err != nil {
			return metadata, err, true
		}
		device, err := q.AdminLockRestrictedDevice(r.Context(), id)
		if err != nil {
			return metadata, err, true
		}
		if !device.UpdatedAt.Time.Equal(body.UpdatedAt) {
			actionConflict(w)
			return metadata, nil, false
		}
		if !device.Blocked {
			return deny(409, "RESTRICTION_CHANGED", "This browser is already unblocked. Reload before continuing")
		}
		metadata.Previous = "blocked"
		metadata.Value = "unblocked"
		return metadata, q.AdminUnblockDevice(r.Context(), db.AdminUnblockDeviceParams{ID: id, Reason: reason, UpdatedBy: actor}), true
	}
	if action == "person.restriction" {
		target, err := q.AdminUserLock(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return metadata, nil, false
		}
		if !target.UpdatedAt.Time.Equal(body.UpdatedAt) {
			actionConflict(w)
			return metadata, nil, false
		}
		if *body.Blocked {
			protected, e := q.HasActiveAdminGrant(r.Context(), id)
			if e != nil {
				return metadata, e, true
			}
			if protected || id == actor {
				return deny(409, "ADMIN_RECOVERY_REQUIRED", "Revoke this account's administrator access before suspending it")
			}
		}
		previous, err := q.IsRestrictedAccount(r.Context(), id)
		if err != nil {
			return metadata, err, true
		}
		if previous == *body.Blocked {
			return deny(409, "RESTRICTION_CHANGED", "This account is already in that state. Reload before continuing")
		}
		metadata.Previous = strconv.FormatBool(previous)
		metadata.Value = strconv.FormatBool(*body.Blocked)
		if err = q.AdminSetAccountRestriction(r.Context(), db.AdminSetAccountRestrictionParams{UserID: id, Blocked: *body.Blocked, Reason: reason, UpdatedBy: actor}); err != nil {
			return metadata, err, true
		}
		if *body.Blocked {
			ids := []pgtype.UUID{id}
			if err = q.RetireRestrictedAccountSessions(r.Context(), ids); err != nil {
				return metadata, err, true
			}
			if err = q.RetireRestrictedAccountGrants(r.Context(), ids); err != nil {
				return metadata, err, true
			}
			if err = q.RevokeUserOAuthExchanges(r.Context(), id); err != nil {
				return metadata, err, true
			}
		}
		return metadata, q.AdminTouchPerson(r.Context(), id), true
	}
	device, err := q.AdminRecognizedDevice(r.Context(), db.AdminRecognizedDeviceParams{UserID: id, AdminReference: *body.DeviceReference})
	if err != nil {
		detailError(w, err)
		return metadata, nil, false
	}
	if err = q.LockRestrictionDevice(r.Context(), device.TokenHash); err != nil {
		return metadata, err, true
	}
	protected, err := q.AdminDeviceHasAdministrator(r.Context(), device.TokenHash)
	if err != nil {
		return metadata, err, true
	}
	targetAdmin, err := q.HasActiveAdminGrant(r.Context(), id)
	if err != nil {
		return metadata, err, true
	}
	if protected || targetAdmin {
		return deny(409, "ADMIN_RECOVERY_REQUIRED", "This browser is associated with an administrator. Review its admin access first")
	}
	ids, err := q.AdminLockDeviceAccounts(r.Context(), db.AdminLockDeviceAccountsParams{TokenHash: device.TokenHash, TargetUserID: id})
	if err != nil {
		return metadata, err, true
	}
	if len(ids) > 100 {
		return deny(409, "REVIEW_LIMIT", "This browser has more than 100 associated accounts. Suspend the accounts individually for review")
	}
	// All accounts are now locked in UUID order; reading the target cannot invert
	// locks against a simultaneous sign-in or another browser restriction.
	target, err := q.AdminUserLock(r.Context(), id)
	if err != nil {
		return metadata, err, true
	}
	if !target.UpdatedAt.Time.Equal(body.UpdatedAt) {
		actionConflict(w)
		return metadata, nil, false
	}
	blocked, err := q.AdminBlockDevice(r.Context(), db.AdminBlockDeviceParams{TokenHash: device.TokenHash, UserAgent: device.UserAgent, Reason: reason, UpdatedBy: actor})
	if err != nil {
		return metadata, err, true
	}
	if err = q.AdminRestrictDeviceAccounts(r.Context(), db.AdminRestrictDeviceAccountsParams{UserIds: ids, Reason: reason, DeviceID: blocked.ID, ActorID: actor}); err != nil {
		return metadata, err, true
	}
	if err = q.AdminRecordDeviceAccounts(r.Context(), db.AdminRecordDeviceAccountsParams{DeviceID: blocked.ID, UserIds: ids}); err != nil {
		return metadata, err, true
	}
	if err = q.RetireRestrictedAccountSessions(r.Context(), ids); err != nil {
		return metadata, err, true
	}
	if err = q.RetireRestrictedAccountGrants(r.Context(), ids); err != nil {
		return metadata, err, true
	}
	metadata.DeviceReference = blocked.Reference
	metadata.AffectedAccounts = len(ids)
	metadata.Value = "blocked"
	return metadata, q.AdminTouchRestrictedAccounts(r.Context(), ids), true
}
