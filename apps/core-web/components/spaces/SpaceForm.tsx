'use client';
import { useId, useRef, useState, type FormEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import type { Space } from '@/types/api';
import { fetchApi } from '@/lib/api';
import { Input } from '@/components/ui/Input';
import { MediaField } from '@hackways/ui/media-field';
import { saveSpaceImage, type SpaceImageAttempt } from '@/lib/space-upload';
import { Icon } from '@/components/ui/Icon';
import { site } from '@/lib/site';
export function SpaceForm({ space, hostName, hasProfile, canArchive = false, managed = false }: { space?: Space; hostName: string; hasProfile: boolean; canArchive?: boolean; managed?: boolean }) {
  const router = useRouter();
  const formID = useId();
  const [step, setStep] = useState(0);
  const [profile, setProfile] = useState(hasProfile), [busy, setBusy] = useState(''), [error, setError] = useState(''), [archive, setArchive] = useState(false);
  const attempt = useRef<{ payload: string; key: string } | null>(null);
  const saved = useRef<Space | undefined>(space);
  const savedFields = useRef('');
  const imageAttempts = useRef<Partial<Record<'logo' | 'banner', SpaceImageAttempt>>>({});
  const [images, setImages] = useState<Record<'logo' | 'banner', { file: File | null; removed: boolean }>>({ logo: { file: null, removed: false }, banner: { file: null, removed: false } });
  async function reloadLatest() {
    if (!saved.current || busy) return; setBusy('reload');
    try { const result = await fetchApi<{ data: Space }>(`/spaces/${saved.current.id}/manage`, { cache: 'no-store' }); saved.current = result.data; savedFields.current = ''; setError('Latest saved version loaded. Your unsaved entries are preserved; review them before saving.'); }
    catch (failure) { setError(failure instanceof Error ? failure.message : 'The latest Space could not be loaded. Your entries are still here.'); }
    finally { setBusy(''); }
  }
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); if (busy) return;
    if (step === 0) { setError(''); setStep(1); return; }
    setBusy('save'); setError('');
    const form = new FormData(e.currentTarget);
    const fields = { name: String(form.get('name')).trim(), description: String(form.get('description')).trim(), guidelines: String(form.get('guidelines')).trim(), website_url: String(form.get('website')).trim() };
    const signature = JSON.stringify(fields);
    const payload = JSON.stringify({ ...fields, ...(saved.current ? { revision: saved.current.revision } : {}) });
    if (attempt.current?.payload !== payload) attempt.current = { payload, key: crypto.randomUUID() };
    try {
      if (!profile) { await fetchApi('/events/organizer/profile', { method: 'POST', body: JSON.stringify({ display_name: hostName }) }); setProfile(true); }
      if (savedFields.current !== signature) {
        const result = await fetchApi<{ data: Space }>(saved.current ? `/spaces/${saved.current.id}` : '/spaces', { method: saved.current ? 'PUT' : 'POST', headers: { 'Idempotency-Key': attempt.current.key }, body: payload });
        saved.current = result.data; savedFields.current = signature;
      }
      if (!saved.current) throw new Error('The Space could not be saved. Retry to continue.');
      for (const kind of ['logo', 'banner'] as const) {
        const image = images[kind];
        if (image.file || image.removed && saved.current[`${kind}_url`]) {
          saved.current = await saveSpaceImage(saved.current, kind, image.file, imageAttempts.current[kind], value => { imageAttempts.current[kind] = value; });
          setImages(current => ({ ...current, [kind]: { file: null, removed: false } }));
        }
      }
      if (managed) { window.location.assign(`${site.organizer}/spaces/${saved.current.slug}`); return; } router.push(`/spaces/${saved.current.slug}`); router.refresh();
    } catch (err) { setError(err instanceof Error ? err.message : 'Your Space could not be saved. Your entries are still here.'); }
    finally { setBusy(''); }
  }
  async function status() {
    if (!space) return; setBusy('status'); setError('');
    try { await fetchApi(`/spaces/${space.id}/status`, { method: 'POST', body: JSON.stringify({ status: space.status === 'active' ? 'archived' : 'active', revision: saved.current?.revision || space.revision }) }); window.location.assign(`${site.organizer}/spaces/${space.slug}`); }
    catch (err) { setError(err instanceof Error ? err.message : 'The Space status could not be updated.'); }
    finally { setBusy(''); }
  }
  return <div className="space-form-page">
    <div className="space-creation-toolbar"><div><Link className="quiet-link space-studio-back" href={space ? `/spaces/${space.slug}` : '/spaces'}><Icon name="arrow" size={15} />{space ? 'Back to Space' : 'All Spaces'}</Link><h1>{space ? 'Edit Space' : 'Create a Space'}</h1></div><div className="space-step-actions">{step === 1 && <button type="button" className="quiet-link" disabled={!!busy} onClick={() => setStep(0)}>Back</button>}<button form={formID} className="app-button app-button-primary" disabled={!!busy}>{busy === 'save' ? 'Saving...' : step === 0 ? 'Continue' : space ? 'Save changes' : 'Create Space'}<Icon name="arrow" size={17} /></button></div></div>
    <ol className="space-creation-steps" aria-label="Creation progress"><li aria-current={step === 0 ? 'step' : undefined}>Community details</li><li aria-hidden="true"><Icon name="arrow" size={15} /></li><li aria-current={step === 1 ? 'step' : undefined}>Appearance <span>(optional)</span></li></ol>
    <form id={formID} className="space-composer" onSubmit={submit} onInvalid={() => setStep(0)}>
      <fieldset className="space-composer-fields" disabled={!!busy}>
        <legend className="visually-hidden">Create your community page</legend>
        <section className="space-details-step" hidden={step !== 0}>
          <Input label="Space name" name="name" defaultValue={space?.name || ''} maxLength={120} placeholder="Your community's name" required />
          <div className="field"><label htmlFor="space-about">About your Space</label><textarea className="app-input" id="space-about" name="description" defaultValue={space?.description || ''} rows={4} maxLength={6000} placeholder="What is your community about?" required /></div>
          <Input label="Website (optional)" name="website" type="url" pattern="https://.*" defaultValue={space?.website_url || ''} placeholder="https://" maxLength={2048} />
          <details className="space-studio-guidelines" open={Boolean(space?.guidelines)}><summary>Community guidelines <span>(optional)</span><Icon name="down" size={17} /></summary><div className="field"><label className="visually-hidden" htmlFor="space-guidelines">Community guidelines</label><textarea className="app-input" id="space-guidelines" name="guidelines" defaultValue={space?.guidelines || ''} rows={3} maxLength={10000} placeholder="Share any values or expectations for your community." /></div></details>
        </section>
        <section className="space-appearance-step" hidden={step !== 1}><p className="space-appearance-intro">Add a cover and logo, or create your Space now and add them later.</p><div className="space-brand-fields">
          {(['banner', 'logo'] as const).map(kind => <MediaField key={kind} label={kind === 'logo' ? 'Logo' : 'Cover image'} variant={kind} file={images[kind].file} initialUrl={saved.current?.[`${kind}_url`]} removed={images[kind].removed} disabled={!!busy} onChange={(file, removed) => setImages(current => ({ ...current, [kind]: { file, removed } }))} />)}
        </div></section>
      </fieldset>
      {error && <div className="journey-error" role="alert"><p>{error}</p>{saved.current && <button className="quiet-link" type="button" disabled={!!busy} onClick={() => void reloadLatest()}>Reload the latest Space</button>}</div>}
    </form>
    {space && canArchive && <section className="space-archive-settings"><h2>{space.status === 'active' ? 'Pause this community.' : 'Welcome people back.'}</h2><p>{space.status === 'active' ? 'Archiving pauses new members and events. Existing events and memberships stay in place.' : 'Reactivating lets people join and lets you organize new events again.'}</p>{space.status === 'active' && !archive ? <button className="quiet-link" onClick={() => setArchive(true)}>Archive this Space</button> : <div className="journey-actions"><button className="app-button app-button-outline" disabled={!!busy} onClick={() => void status()}>{busy === 'status' ? 'Saving...' : space.status === 'active' ? 'Confirm archive' : 'Reactivate Space'}</button>{archive && <button className="quiet-link" disabled={!!busy} onClick={() => setArchive(false)}>Keep it active</button>}</div>}</section>}
  </div>;
}
