"use client";
import { useRef, useState, type FormEvent } from "react";
import { Checkbox } from "@hackways/ui/checkbox";
import type { EditableGuest } from "@/lib/guests";
import { ApiClientError } from "@/lib/api";

import { decisionBody, saveGuestDecision, type GuestDecision } from '@/lib/guest-decisions';
import { Icon } from '@/components/ui/Icon';

export function GuestDecisionForm({
  eventId,
  guest,
  canApprove,
  onBusy,
  onUpdated,
  onReload,
}: {
  eventId: string;
  guest: EditableGuest;
  canApprove: boolean;
  onBusy: (busy: boolean) => void;
  onUpdated: (message: string) => void;
  onReload: () => void;
}) {
  const lock = useRef(false),
    attempt = useRef<{ body: string; key: string } | null>(null);
  const [notify, setNotify] = useState(!!guest.guest_email),
    [message, setMessage] = useState("");
  const [busy, setBusy] = useState(""),
    [error, setError] = useState(""),
    [conflict, setConflict] = useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const button = (event.nativeEvent as SubmitEvent).submitter;
    const status = button instanceof HTMLButtonElement ? button.value : "";
    if (
      lock.current ||
      !["confirmed", "rejected"].includes(status) ||
      (status === "confirmed" && !canApprove)
    )
      return;
    const body = decisionBody(guest, status as GuestDecision, notify, message);
    if (attempt.current?.body !== body)
      attempt.current = { body, key: crypto.randomUUID() };
    lock.current = true;
    setBusy(status);
    onBusy(true);
    setError("");
    setConflict(false);
    try {
      onUpdated(await saveGuestDecision(eventId, guest, status as GuestDecision, attempt.current));
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "The decision could not be saved.",
      );
      setConflict(
        err instanceof ApiClientError &&
          [
            "REGISTRATION_CHANGED",
            "REGISTRATION_EXPIRED",
            "STATUS_CHANGE_UNAVAILABLE",
          ].includes(err.code),
      );
    } finally {
      lock.current = false;
      setBusy("");
      onBusy(false);
    }
  }
  return (
    <section className="guest-quick-decision" aria-label="Review application">
      <p className="guest-pending-note">This guest is awaiting approval.</p>
      <form onSubmit={submit} aria-busy={!!busy}>
        <label className="guest-status-notify">
          <Checkbox
            checked={notify}
            onCheckedChange={setNotify}
            disabled={!!busy || !guest.guest_email}
          />
          <span>Notify guest</span>
        </label>
        {notify && (
          <div className="field">
            <label className="visually-hidden" htmlFor="quick-decision-message">
              Optional message for the guest
            </label>
            <textarea
              id="quick-decision-message"
              className="app-input"
              placeholder="Add an optional message…"
              rows={3}
              maxLength={2000}
              value={message}
              onChange={(event) => setMessage(event.target.value)}
              disabled={!!busy}
            />
          </div>
        )}
        {error && (
          <div className="journey-error" role="alert">
            <p>{error}</p>
            {conflict && (
              <button type="button" className="quiet-link" onClick={onReload}>
                Load current guest details
              </button>
            )}
          </div>
        )}
        <div className="guest-decision-buttons">
          <button
            className="app-button guest-approve-button"
            type="submit"
            name="status"
            value="confirmed"
            disabled={!!busy || conflict || !canApprove}
          >
            <Icon name="check" size={18} />
            {busy === "confirmed" ? "Approving…" : "Approve"}
          </button>
          <button
            className="app-button guest-decline-button"
            type="submit"
            name="status"
            value="rejected"
            disabled={!!busy || conflict}
          >
            <Icon name="close" size={18} />
            {busy === "rejected" ? "Declining…" : "Decline"}
          </button>
        </div>
      </form>
    </section>
  );
}
