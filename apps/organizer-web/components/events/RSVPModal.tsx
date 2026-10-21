"use client";

import React, { useState } from "react";
import { Event, Registration } from "@/types/api";
import { Modal } from "@/components/ui/Modal";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { fetchApi, ApiClientError } from "@/lib/api";

interface RSVPModalProps {
  event: Event | null;
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (eventId: string) => void;
}

export const RSVPModal: React.FC<RSVPModalProps> = ({
  event,
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [notes, setNotes] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [confirmedRSVP, setConfirmedRSVP] = useState<Registration | null>(null);

  if (!event) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) {
      setError("Email address is required.");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await fetchApi<{ data: Registration; message: string }>(
        "/rsvps",
        {
          method: "POST",
          body: JSON.stringify({
            event_id: event.id,
            guest_name: name.trim(),
            guest_email: email.trim(),
            notes: notes.trim(),
          }),
        }
      );

      setConfirmedRSVP(response.data);
      onSuccess(event.id);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError("Failed to reserve your ticket. Please try again.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setName("");
    setEmail("");
    setNotes("");
    setError(null);
    setConfirmedRSVP(null);
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={confirmedRSVP ? "Registration Confirmed!" : `RSVP: ${event.title}`}
      description={
        confirmedRSVP
          ? "Your spot has been secured. A confirmation email has been dispatched."
          : `Date: ${new Date(event.starts_at).toLocaleDateString(undefined, {
              month: "short",
              day: "numeric",
              year: "numeric",
            })}`
      }
    >
      {confirmedRSVP ? (
        <div className="py-4 space-y-4 text-center">
          <div className="w-16 h-16 mx-auto bg-emerald-950/80 border border-emerald-800 rounded-full flex items-center justify-center text-emerald-400">
            <svg
              className="w-8 h-8"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>

          <div>
            <Badge variant="success">Confirmed Ticket</Badge>
            <p className="mt-2 text-sm text-[#475069] font-medium">
              Attendee: {confirmedRSVP.guest_name || confirmedRSVP.guest_email}
            </p>
            <p className="text-xs text-[#757F9A] mt-1 font-mono">
              RSVP ID: {confirmedRSVP.id}
            </p>
          </div>

          <div className="pt-4 border-t border-[#E4E7F0]">
            <Button variant="primary" size="md" className="w-full" onClick={handleClose}>
              Done
            </Button>
          </div>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          {error && (
            <div className="p-3 bg-rose-950/60 border border-rose-200/80 rounded-xl text-xs text-rose-800">
              {error}
            </div>
          )}

          <Input
            label="Full Name"
            placeholder="Ada Lovelace"
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={isLoading}
          />

          <Input
            label="Email Address"
            type="email"
            placeholder="ada@example.com"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={isLoading}
          />

          <div>
            <label htmlFor="rsvp-notes" className="mb-1.5 block text-sm font-semibold text-[#475069]">
              Special Notes / Dietary / Accessibility (Optional)
            </label>
            <textarea
              id="rsvp-notes"
              className="w-full rounded-xl border border-[#E4E7F0] bg-white px-4 py-3 text-sm text-[#0E121E] placeholder:text-[#757F9A] transition focus:border-[#34349C] focus:outline-none focus:ring-2 focus:ring-[#34349C] disabled:opacity-50"
              rows={2}
              placeholder="Any requirements or comments for the organizer..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              disabled={isLoading}
            />
          </div>

          <div className="pt-3 border-t border-[#E4E7F0] flex justify-end gap-3">
            <Button
              type="button"
              variant="outline"
              size="md"
              onClick={handleClose}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="md"
              isLoading={isLoading}
            >
              Confirm Reservation
            </Button>
          </div>
        </form>
      )}
    </Modal>
  );
};
