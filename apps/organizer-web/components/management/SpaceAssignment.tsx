'use client';
import { useEffect, useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import type { Space } from '@/types/api';
import type { ManagedEvent } from '@/lib/management';
import { fetchApi } from '@/lib/api';
import { site } from '@/lib/site';
import { Icon } from '@/components/ui/Icon';
export function SpaceAssignment({ event, spaces, current }: { event: ManagedEvent; spaces: Space[]; current: Space | null }) {
  const router = useRouter(), menu = useRef<HTMLDetailsElement>(null), lock = useRef(false);
  const [slug, setSlug] = useState(current?.slug || ''), [busy, setBusy] = useState(false), [error, setError] = useState('');
  const editable = event.status !== 'cancelled' && Date.parse(event.starts_at) > Date.now();
  const choices = current?.status === 'active' && !spaces.some(space => space.slug === current.slug) ? [current, ...spaces] : spaces;
  useEffect(() => {
    function close(e: PointerEvent) { if (e.target instanceof Node && !menu.current?.contains(e.target)) menu.current?.removeAttribute('open'); }
    function escape(e: KeyboardEvent) { if (e.key === 'Escape' && menu.current?.open) { menu.current.open = false; menu.current.querySelector('summary')?.focus(); } }
    document.addEventListener('pointerdown', close); document.addEventListener('keydown', escape);
    return () => { document.removeEventListener('pointerdown', close); document.removeEventListener('keydown', escape); };
  }, []);
  function choose(value: string) { setSlug(value); if (menu.current) { menu.current.open = false; menu.current.querySelector('summary')?.focus(); } }
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); if (lock.current) return; lock.current = true; setBusy(true); setError('');
    try { await fetchApi(`/events/${event.id}/space`, { method: 'PUT', body: JSON.stringify({ space_slug: slug, revision: event.revision }) }); router.refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : 'The event’s Space could not be changed.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <section><h2>A home for this event.</h2><p>{current ? <>Part of <a className="quiet-link" href={`${site.app}/spaces/${current.slug}`}>{current.name}</a>.</> : 'This is a personal event.'}</p>{editable && <form className="journey-form" onSubmit={submit}><div className="field"><span>Organize under</span><details className="host-space-menu" ref={menu}><summary onClick={e => { if (busy) e.preventDefault(); }} aria-disabled={busy || undefined}><Icon name="people" size={18} /><span>{choices.find(space => space.slug === slug)?.name || (current?.slug === slug ? `${current.name} (archived)` : 'Personal event')}</span><Icon name="down" size={15} /></summary><div><button type="button" onClick={() => choose('')} aria-pressed={!slug}>Personal event{!slug && <Icon name="check" size={16} />}</button>{choices.filter(space => space.status === 'active').map(space => <button type="button" key={space.slug} aria-pressed={slug === space.slug} onClick={() => choose(space.slug)}>{space.name}{slug === space.slug && <Icon name="check" size={16} />}</button>)}</div></details></div><button className="app-button app-button-outline" disabled={busy || slug === (current?.slug || '')}>{busy ? 'Saving…' : 'Save Space'}</button>{error && <p className="journey-error" role="alert">{error}</p>}</form>}<a className="quiet-link" href={`${site.app}/spaces?view=owned`}>My Spaces</a></section>;
}
