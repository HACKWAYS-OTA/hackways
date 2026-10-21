'use client';
import { forwardRef, useEffect, useImperativeHandle, useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import type { ManagedEvent } from '@/lib/management';
import type { Event } from '@/types/api';
import { fetchApi } from '@/lib/api';
import { safeLink } from '@/lib/dashboard';
import { appearanceOf, type Appearance } from '@/lib/appearance';
import { uploadEventCover, type CoverAttempt } from '@/lib/cover-upload';
import { CoverPicker } from '@/components/hosting/CoverPicker';
import { ThemePicker } from '@/components/hosting/ThemePicker';
import { useEventEditSession } from './EventEditSession';
export interface EventAppearanceHandle { save: () => Promise<void> }
interface AppearanceEditorProps { event: ManagedEvent; compact?: boolean; onError?: (message: string) => void; onPreview?: (appearance: Appearance, file: File | null, remove: boolean) => void }
export const EventAppearanceEditor = forwardRef<EventAppearanceHandle, AppearanceEditorProps>(function EventAppearanceEditor({ event, compact = false, onPreview, onError }, ref) {
  const session = useEventEditSession();
  const router = useRouter(), current = useRef<Event>(event), attempt = useRef<CoverAttempt | null>(null), lock = useRef(false);
  const [appearance, setAppearance] = useState(() => appearanceOf(event)), [file, setFile] = useState<File | null>(null), [remove, setRemove] = useState(false), [busy, setBusy] = useState(''), [error, setError] = useState(''), [saved, setSaved] = useState(false);
  useEffect(() => { onPreview?.(appearance, file, remove); }, [appearance, file, remove, onPreview]);
  async function persist() {
    if (lock.current) throw new Error('Appearance is already being saved.');
    if (session) current.current = session.latest();
    lock.current = true; setBusy('Saving appearance…'); setError(''); setSaved(false);
    try {
      if (file) { setBusy('Uploading banner…'); current.current = await uploadEventCover(file, current.current, attempt.current, next => { attempt.current = next; }); session?.update(current.current); }
      setBusy('Saving appearance…');
      const result = await fetchApi<{ data: Event }>(`/events/${event.id}/appearance`, { method: 'PUT', body: JSON.stringify({ ...appearance, revision: current.current.revision, remove_banner: remove && !file }) });
      current.current = result.data; session?.update(result.data); setFile(null); setRemove(false); attempt.current = null; setSaved(true);
    } catch (err) { setError(err instanceof Error ? err.message : 'Appearance could not be saved. Your selection is still here.'); throw err; }
    finally { lock.current = false; setBusy(''); }
  }
  useImperativeHandle(ref, () => ({ save: persist }));
  async function save(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); if (lock.current || (session && !session.begin('appearance'))) return;
    try { await persist(); router.refresh(); }
    catch { /* persist retains the selection and exposes the actual failure. */ }
    finally { session?.finish(); }
  }
  const disabled = !!busy || !!session?.busyArea;
  return <form className={`host-appearance-editor${compact ? ' is-compact' : ''}`} onSubmit={save} aria-busy={!!busy}><div className="host-appearance-picker"><CoverPicker onError={onError} file={file} initialUrl={remove ? '' : safeLink(event.banner_url) || ''} onRemoveExisting={() => { setRemove(true); setSaved(false); }} onChange={next => { setSaved(false); setFile(next); if (next) setRemove(false); attempt.current = null; }} disabled={disabled} />{remove && <button type="button" className="quiet-link" disabled={disabled} onClick={() => { setRemove(false); setSaved(false); }}>Keep current banner</button>}<ThemePicker value={appearance} onChange={next => { setAppearance(next); setSaved(false); }} disabled={disabled} /></div><div className="host-appearance-copy">{!compact && <h2>Make it feel like your event.</h2>}{!compact && <><p>Choose a cover and a theme for your event page. Your guests will see the saved look.</p><p>Ambient uses your banner as a soft background. Light and Dark are available for every theme.</p><p className="field-help">JPG, PNG or WebP, up to 5 MB. The current banner stays in place until a replacement is verified.</p></>}{error && !compact && <div className="journey-error" role="alert"><p>{error}</p><p>Any confirmed banner upload is retained. Reload the event if its revision changed before trying again.</p>{session && <button type="button" className="quiet-link" disabled={disabled} onClick={() => session.reload().catch(err => setError(err instanceof Error ? err.message : 'The event could not be reloaded.'))}>Reload saved event (discard edits)</button>}</div>}{saved && !compact && <p className="journey-notice" role="status">Your event’s appearance is saved.</p>}{!compact && <button className="app-button app-button-primary" disabled={disabled}>{busy || 'Save appearance'}</button>}</div></form>;
});
