'use client';
import { useRef, useState, type FormEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import type { Space } from '@/types/api';
import { fetchApi } from '@/lib/api';
import { Input } from '@/components/ui/Input';
export function SpaceForm({ space, hostName, hasProfile }: { space?: Space; hostName: string; hasProfile: boolean }) {
  const router = useRouter();
  const [profile, setProfile] = useState(hasProfile), [busy, setBusy] = useState(''), [error, setError] = useState(''), [archive, setArchive] = useState(false);
  const attempt = useRef<{ payload: string; key: string } | null>(null);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); if (busy) return; setBusy('save'); setError('');
    const form = new FormData(e.currentTarget);
    const payload = JSON.stringify({ name: String(form.get('name')).trim(), description: String(form.get('description')).trim(), guidelines: String(form.get('guidelines')).trim(), website_url: String(form.get('website')).trim(), ...(space ? { revision: space.revision } : {}) });
    if (attempt.current?.payload !== payload) attempt.current = { payload, key: crypto.randomUUID() };
    try {
      if (!profile) { await fetchApi('/events/organizer/profile', { method: 'POST', body: JSON.stringify({ display_name: hostName }) }); setProfile(true); }
      const result = await fetchApi<{ data: Space }>(space ? `/spaces/${space.id}` : '/spaces', { method: space ? 'PUT' : 'POST', headers: { 'Idempotency-Key': attempt.current.key }, body: payload });
      router.push(`/spaces/${result.data.slug}`); router.refresh();
    } catch (err) { setError(err instanceof Error ? err.message : 'Your Space could not be saved. Your entries are still here.'); }
    finally { setBusy(''); }
  }
  async function status() {
    if (!space) return; setBusy('status'); setError('');
    try { await fetchApi(`/spaces/${space.id}/status`, { method: 'POST', body: JSON.stringify({ status: space.status === 'active' ? 'archived' : 'active', revision: space.revision }) }); router.push(`/spaces/${space.slug}`); router.refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : 'The Space status could not be updated.'); }
    finally { setBusy(''); }
  }
  return <div className="space-form-page"><Link className="quiet-link" href={space ? `/spaces/${space.slug}` : '/spaces'}>{space ? 'Back to Space' : 'Back to Spaces'}</Link><header className="desk-heading"><div><h1>{space ? 'Make this Space your own.' : 'Start a Space.'}</h1><p>{space ? 'Keep your community’s details up to date.' : 'Bring people together around something you care about.'}</p></div></header><form className="journey-form" onSubmit={submit}><fieldset className="space-form-fields" disabled={!!busy}><legend className="visually-hidden">Community details</legend><Input label="Space name" name="name" defaultValue={space?.name || ''} maxLength={120} required /><div className="field"><label htmlFor="space-about">What’s this community about?</label><textarea className="app-input" id="space-about" name="description" defaultValue={space?.description || ''} rows={4} maxLength={6000} required /></div><div className="field"><label htmlFor="space-guidelines">Community guidelines (optional)</label><textarea className="app-input" id="space-guidelines" name="guidelines" defaultValue={space?.guidelines || ''} rows={3} maxLength={10000} /></div><Input label="Website (optional)" name="website" type="url" pattern="https://.*" defaultValue={space?.website_url || ''} maxLength={2048} helperText="Use a secure https:// link." /></fieldset>{error && <div className="journey-error" role="alert"><p>{error}</p>{space && <button className="quiet-link" type="button" onClick={() => router.refresh()}>Reload the latest Space</button>}</div>}<div className="space-save"><p>You own this Space. Members can join and take part in events; organizing stays with you.</p><button className="app-button app-button-primary" disabled={!!busy}>{busy === 'save' ? 'Saving…' : space ? 'Save changes' : 'Create Space'}</button></div></form>{space && <section className="space-archive-settings"><h2>{space.status === 'active' ? 'Take a pause.' : 'Welcome people back.'}</h2><p>{space.status === 'active' ? 'Archiving pauses new members and events. Existing events and memberships stay in place.' : 'Reactivating lets people join and lets you organize new events again.'}</p>{space.status === 'active' && !archive ? <button className="quiet-link" onClick={() => setArchive(true)}>Archive this Space</button> : <div className="journey-actions"><button className="app-button app-button-outline" disabled={!!busy} onClick={() => void status()}>{busy === 'status' ? 'Saving…' : space.status === 'active' ? 'Confirm archive' : 'Reactivate Space'}</button>{archive && <button className="quiet-link" disabled={!!busy} onClick={() => setArchive(false)}>Keep it active</button>}</div>}</section>}</div>;
}
