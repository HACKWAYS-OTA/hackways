'use client';

import { createContext, useContext, useRef, useState, type ReactNode } from 'react';
import type { Event } from '@/types/api';
import type { ManagedEvent } from '@/lib/management';
import { fetchApi } from '@/lib/api';

type SaveArea = 'details' | 'appearance' | 'reload';
interface EditSession {
  event: ManagedEvent;
  generation: number;
  busyArea: SaveArea | null;
  timeZones: readonly string[];
  latest: () => ManagedEvent;
  begin: (area: SaveArea) => boolean;
  finish: () => void;
  update: (event: Event) => void;
  reload: () => Promise<void>;
}
const Session = createContext<EditSession | null>(null);
export function useEventEditSession() { return useContext(Session); }

/** Coordinates the existing detail and appearance endpoints without remounting dirty forms. */
export function EventEditSession({ initial, timeZones, children }: { initial: ManagedEvent; timeZones: readonly string[]; children: ReactNode }) {
  const latest = useRef(initial), saving = useRef(false);
  const [event, setEvent] = useState(initial), [generation, setGeneration] = useState(0);
  const [busyArea, setBusyArea] = useState<SaveArea | null>(null);
  function begin(area: SaveArea) {
    if (saving.current) return false;
    saving.current = true; setBusyArea(area); return true;
  }
  function finish() { saving.current = false; setBusyArea(null); }
  function update(next: Event) {
    if (next.id !== latest.current.id || typeof next.revision !== 'number' || !Number.isSafeInteger(next.revision)) throw new Error('The saved event response is incomplete. Reload the saved event before continuing.');
    if (next.revision < latest.current.revision) return;
    const previous = latest.current;
    latest.current = { ...previous, ...next, revision: next.revision,
      held_rsvp_count: next.held_rsvp_count ?? previous.held_rsvp_count,
      registration_mode: next.registration_mode ?? previous.registration_mode,
      allow_guest: next.allow_guest ?? previous.allow_guest,
      waitlist_enabled: next.waitlist_enabled ?? previous.waitlist_enabled,
      offer_minutes: next.offer_minutes ?? previous.offer_minutes,
    };
    setEvent(latest.current);
  }
  async function reload() {
    if (!begin('reload')) return;
    try {
      const result = await fetchApi<{ data: ManagedEvent }>(`/events/${latest.current.id}/manage`, { cache: 'no-store' });
      update(result.data); setGeneration(value => value + 1);
    } finally { finish(); }
  }
  return <Session.Provider value={{ event, generation, busyArea, timeZones, latest: () => latest.current, begin, finish, update, reload }}>{children}</Session.Provider>;
}
