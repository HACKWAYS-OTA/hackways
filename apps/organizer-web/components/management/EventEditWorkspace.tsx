'use client';

import type { ManagedEvent } from '@/lib/management';
import { useCallback, useRef, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { appearanceOf, appearanceStyle, type Appearance } from '@/lib/appearance';
import { AmbientBackdrop } from '@/components/hosting/AmbientBackdrop';
import { EventEditSession, useEventEditSession } from './EventEditSession';
import { EventAppearanceEditor, type EventAppearanceHandle } from './EventAppearanceEditor';
import { EventSettings } from './EventSettings';
import { EventFeedbackDialog } from './EventFeedbackDialog';
import { Icon } from '@/components/ui/Icon';

export function EventEditWorkspace({ event, timeZones }: { event: ManagedEvent; timeZones: readonly string[] }) {
  return <EventEditSession initial={event} timeZones={timeZones}><Workspace /></EventEditSession>;
}
function Workspace() {
  const session = useEventEditSession();
  const router = useRouter();
  const appearanceEditor = useRef<EventAppearanceHandle>(null);
  const [feedback, setFeedback] = useState<{ kind: 'success' | 'error'; message?: string } | null>(null);
  const [preview, setPreview] = useState<{ appearance: Appearance; file: File | null; remove: boolean }>(() => ({ appearance: appearanceOf(session?.event || {}), file: null, remove: false }));
  const updatePreview = useCallback((appearance: Appearance, file: File | null, remove: boolean) => setPreview({ appearance, file, remove }), []);
  if (!session) return null;
  return <div className={`event-edit-page event-edit-preview event-editor-theme event-theme event-display-${preview.appearance.theme_display}`} style={appearanceStyle(preview.appearance)}>
    <div className={`event-edit-canvas event-theme theme-${preview.appearance.theme} event-display-${preview.appearance.theme_display}`} style={appearanceStyle(preview.appearance)} aria-hidden="true"><AmbientBackdrop active={preview.appearance.theme === 'ambient'} file={preview.file} source={preview.remove ? '' : session.event.banner_url} /></div>
    <header className="event-edit-toolbar"><Link className="quiet-link" href={`/events/${session.event.slug}`}><Icon name="arrow" className="reverse-icon" size={16} />Event overview</Link><h1 className="visually-hidden">Edit event</h1><div><button type="button" className="app-button app-button-outline" disabled={!!session.busyArea} onClick={() => router.push(`/events/${session.event.slug}`)}>Discard changes</button><button type="submit" form="event-edit-details-form" className="app-button app-button-primary" disabled={!!session.busyArea}>{session.busyArea && session.busyArea !== 'reload' ? 'Saving…' : 'Save changes'}</button></div></header>
    <div className="event-edit-studio" data-display={preview.appearance.theme_display} data-style={preview.appearance.theme_style} style={appearanceStyle(preview.appearance)}>
    <div className="event-edit-art"><EventAppearanceEditor ref={appearanceEditor} key={`appearance-${session.generation}`} event={session.event} compact onError={message => setFeedback({ kind: 'error', message })} onPreview={updatePreview} /></div>
    <EventSettings key={`details-${session.generation}`} event={session.event} onSaved={() => setFeedback({ kind: 'success' })} onError={message => setFeedback({ kind: 'error', message })} saveAppearance={async () => { if (!appearanceEditor.current) throw new Error('The appearance editor is not ready. Please try again.'); await appearanceEditor.current.save(); }} />
  </div><EventFeedbackDialog feedback={feedback} onClose={() => setFeedback(null)} /></div>;
}
