'use client';
import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import type { Space } from '@/types/api';
import { fetchApi } from '@/lib/api';
export function SpaceMembership({ space, membership, signedIn }: { space: Space; membership: { owned: boolean; joined: boolean } | null; signedIn: boolean }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [confirm, setConfirm] = useState(false);
  async function change() {
    setBusy(true); setError('');
    try { await fetchApi(`/spaces/${space.id}/${membership?.joined ? 'leave' : 'join'}`, { method: 'POST' }); setConfirm(false); router.refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : 'Your membership could not be updated.'); }
    finally { setBusy(false); }
  }
  return <div className="space-membership-actions">{membership?.owned ? <><span className="space-owner-label">Your Space</span>{space.status === 'active' && <Link className="app-button app-button-primary" href={`/create?space=${space.slug}`}>Create an event</Link>}<Link className="quiet-link" href={`/spaces/${space.slug}/settings`}>Edit Space</Link></> : membership?.joined ? <><span className="space-owner-label">You’re a member</span><button type="button" className="quiet-link" disabled={busy} onClick={() => setConfirm(!confirm)}>Leave Space</button>{confirm && <div className="space-leave-confirm"><p>Leave this community? Your event registrations will stay as they are.</p><button className="app-button app-button-outline" disabled={busy} onClick={() => void change()}>{busy ? 'Leaving…' : 'Yes, leave Space'}</button></div>}</> : space.status === 'archived' ? <span className="space-owner-label">Membership is paused</span> : signedIn ? <button className="app-button app-button-primary" disabled={busy} onClick={() => void change()}>{busy ? 'Joining…' : 'Join Space'}</button> : <Link className="app-button app-button-primary" href={`/login?next=${encodeURIComponent(`/spaces/${space.slug}`)}`}>Join Space</Link>}{error && <p role="alert" className="journey-error">{error}</p>}</div>;
}
