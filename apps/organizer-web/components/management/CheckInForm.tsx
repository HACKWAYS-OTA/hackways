"use client";
import { useRef, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import type { ManagedEvent } from "@/lib/management";
import { fetchApi } from "@/lib/api";
import { Icon } from "@/components/ui/Icon";
import { PassScanner } from "./PassScanner";
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
    [message, setMessage] = useState("");
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
    try {
      await fetchApi(`/events/${event.id}/check-in`, {
        method: "POST",
        body: JSON.stringify({ code: code.trim() }),
      });
      setMessage("Guest checked in. This pass is now marked as used.");
      router.refresh();
      return true;
    } catch (err) {
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
