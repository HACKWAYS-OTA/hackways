"use client";
import { useRef, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import type { ManagedEvent } from "@/lib/management";
import { fetchApi, ApiClientError } from "@/lib/api";
import { Icon } from "@/components/ui/Icon";
import { PassScanner } from "./PassScanner";
import { UndoCheckInButton } from "./UndoCheckInButton";
type Attendee = { ticket_id:string; name:string; email:string; registration_code:string; checked_in_at:string|null };
function scannedAttendee(value:unknown): Attendee|null {
 if(!value || typeof value!=="object" || !("ticket_id" in value) || typeof value.ticket_id!=="string" || !("name" in value) || typeof value.name!=="string" || !("email" in value) || typeof value.email!=="string" || !("registration_code" in value) || typeof value.registration_code!=="string" || !("checked_in_at" in value) || !(typeof value.checked_in_at==="string" || value.checked_in_at===null)) return null;
 return {ticket_id:value.ticket_id,name:value.name,email:value.email,registration_code:value.registration_code,checked_in_at:value.checked_in_at};
}
export function CheckInForm({
  event,
  open,
}: {
  event: ManagedEvent;
  open: boolean;
}) {
  const router = useRouter(),
    lock = useRef(false),
    result = useRef<HTMLDivElement>(null);
  const [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [message, setMessage] = useState(""),
    [attendee,setAttendee] = useState<Attendee|null>(null);
  async function scan(code: string) {
    if (lock.current || !open) return false;
    if (!code.trim() || code.length > 2048) {
      setError("This is not a supported Hackways QR pass.");
      return false;
    }
    lock.current = true;
    setBusy(true);
    setError("");
    setMessage("");
    setAttendee(null);
    try {
      const response = await fetchApi<{attendee:Attendee}>(`/events/${event.id}/check-in`, {
        method: "POST",
        body: JSON.stringify({ code: code.trim() }),
      });
      setAttendee(scannedAttendee(response.attendee));
      setMessage("Checked in successfully.");
      router.refresh();
      return true;
    } catch (err) {
      if(err instanceof ApiClientError && err.code==="ALREADY_CHECKED_IN") setAttendee(scannedAttendee(err.details?.attendee));
      setError(
        err instanceof Error
          ? err.message
          : "Check-in failed. You can retry this pass.",
      );
      return false;
    } finally {
      lock.current = false;
      setBusy(false);
      requestAnimationFrame(() =>
        result.current?.focus({ preventScroll: true }),
      );
    }
  }
  async function manual(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    if (await scan(String(new FormData(form).get("code") || ""))) form.reset();
  }
  return (
    <section className="host-entry scanner-entry">
      <div className="scanner-entry-heading">
        <h2>Scan QR passes</h2>
      </div>
      {!open && (
        <p className="host-entry-closed">
          Check-in opens 24 hours before the event and closes when it ends.
          Publish the event to enable check-in.
        </p>
      )}
      <PassScanner
        disabled={!open || busy}
        onScan={async (code) => {
          await scan(code);
        }}
      />
      <div className="checkin-result" ref={result} tabIndex={-1}>
        {busy && <p role="status">Checking pass...</p>}
        {error && (
          <p className="journey-error" role="alert">
            {error}
          </p>
        )}
        {message && (
          <p className="journey-notice" role="status">
            <Icon name="check" size={18} />
            {message}
          </p>
        )}
        {attendee && <section className="checkin-attendee" aria-label="Scanned attendee"><h3>{attendee.name || 'Attendee'}</h3><p>{attendee.email}</p><p className="field-help">{attendee.registration_code}</p>{attendee.checked_in_at && <><p className="field-help">Checked in {new Date(attendee.checked_in_at).toLocaleString('en-IN',{timeZone:event.timezone})}</p><UndoCheckInButton key={`${attendee.ticket_id}:${attendee.checked_in_at}`} eventId={event.id} ticketId={attendee.ticket_id} checkedInAt={attendee.checked_in_at} name={attendee.name || attendee.email} disabled={event.management_closed} onBusy={setBusy} onUndone={()=>{setAttendee(current=>current?{...current,checked_in_at:null}:null);setError('');setMessage('Check-in undone. The attendee can be scanned again.');}} /></>}</section>}
      </div>
      <details className="manual-pass-entry">
        <summary>Enter a pass reference instead</summary>
        <form onSubmit={manual} aria-busy={busy}>
          <label htmlFor="entry-code">Pass reference</label>
          <div className="host-entry-field">
            <input
              id="entry-code"
              name="code"
              required
              maxLength={2048}
              autoComplete="off"
              autoCapitalize="none"
              spellCheck={false}
              disabled={!open || busy}
            />
            <button
              className="app-button app-button-primary"
              disabled={!open || busy}
            >
              {busy ? "Checking in..." : "Check in"}
              <Icon name="arrow" size={16} />
            </button>
          </div>
          <p className="field-help">
            Or find the guest below and use Check in after verifying their
            identity.
          </p>
        </form>
      </details>
    </section>
  );
}
