'use client';
import { useRef, useState, type FormEvent } from 'react';
import { Checkbox } from '@hackways/ui/checkbox';
import { Select } from '@hackways/ui/select';
import { adminRequest, AdminRequestError, AdminAuthenticationError } from '@/lib/admin-request';
import { ProgramAuthRecovery } from './ProgramAuthRecovery';
type Preview = { count: number; audience_hash: string; html: string; subject: string; recipient: string };
type Draft = { subject: string; body: string; scope: string; ids: string[]; decision: string; search: string };
type Attempt = { key: string; request: Draft & { audience_hash: string; expected_count: number; confirmed: boolean } };
export function ProgramCommunication({ slug, ids, scope, search, decision }: { slug: string; ids: string[]; scope: string; search: string; decision: string }) {
  const [draft, setDraft] = useState<Draft>({ subject: 'An update from {{program}}', body: '', scope: ids.length ? 'selected' : scope === 'filtered' ? 'filtered' : 'all', ids, search, decision });
  const [preview, setPreview] = useState<Preview | null>(null), [confirmed, setConfirmed] = useState(false), [busy, setBusy] = useState(false), [error, setError] = useState<Error | null>(null), [notice, setNotice] = useState(''), [uncertain, setUncertain] = useState(false);
  const attempt = useRef<Attempt | null>(null);
  function change<K extends keyof Draft>(key: K, value: Draft[K]) { setDraft(previous => ({ ...previous, [key]: value })); setPreview(null); setConfirmed(false); setNotice(''); attempt.current = null; }
  async function prepare(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (busy || uncertain) return; setBusy(true); setError(null); setNotice('');
    try { setPreview(await adminRequest<Preview>(`global/programs/${slug}/communications/preview`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(draft) }, true)); setConfirmed(false); }
    catch (cause) { setError(cause instanceof Error ? cause : new Error('The preview could not be prepared. Your message is still here.')); }
    finally { setBusy(false); }
  }
  async function send() {
    if (busy || !preview || !confirmed) return; setBusy(true); setError(null);
    if (!attempt.current) attempt.current = { key: crypto.randomUUID(), request: { ...draft, audience_hash: preview.audience_hash, expected_count: preview.count, confirmed: true } };
    try {
      const result = await adminRequest<{ scheduled: number; reference: number }>(`global/programs/${slug}/communications/send`, { method: 'POST', headers: { 'Content-Type': 'application/json', 'Idempotency-Key': attempt.current.key }, body: JSON.stringify(attempt.current.request) }, true);
      setNotice(`Message ${result.reference} scheduled for ${result.scheduled} ${result.scheduled === 1 ? 'applicant' : 'applicants'}.`); setConfirmed(false); setUncertain(false); setPreview(null); attempt.current = null;
    } catch (cause) {
      setError(cause instanceof Error ? cause : new Error('The send request could not be confirmed. Retry safely.'));
      if (cause instanceof AdminAuthenticationError || (cause instanceof AdminRequestError && cause.status < 500)) { attempt.current = null; setUncertain(false); setPreview(null); setConfirmed(false); }
      else setUncertain(true);
    } finally { setBusy(false); }
  }
  return <div className="program-mail-workspace"><form className="program-mail-compose" onSubmit={event => void prepare(event)}>
    <h2>Write to your applicants.</h2><p>Each email greets its recipient by name and uses the program’s email design.</p>
    <fieldset disabled={busy || uncertain} className="program-mail-fields">
      <label htmlFor="program-mail-audience">Recipients<Select id="program-mail-audience" label="Recipients" value={draft.scope} onValueChange={value => change('scope', value)} options={[...(ids.length ? [{ value: 'selected', label: `${ids.length} selected ${ids.length === 1 ? 'applicant' : 'applicants'}` }] : []), { value: 'all', label: 'All submitted applicants' }, { value: 'filtered', label: 'A filtered group' }]} /></label>
      {draft.scope === 'filtered' && <div className="program-admin-field-pair"><label htmlFor="program-mail-decision">Review group<Select id="program-mail-decision" label="Review group" value={draft.decision} onValueChange={value => change('decision', value)} options={[{ value: '', label: 'All decisions' }, { value: 'pending', label: 'Needs review' }, { value: 'accepted', label: 'Accepted' }, { value: 'waitlisted', label: 'Waitlisted' }, { value: 'rejected', label: 'Not selected' }]} /></label><label htmlFor="program-mail-search">Name, email or city<input id="program-mail-search" maxLength={200} value={draft.search} onChange={event => change('search', event.target.value)} /></label></div>}
      <label htmlFor="program-mail-subject">Subject<input id="program-mail-subject" required minLength={2} maxLength={200} value={draft.subject} onChange={event => change('subject', event.target.value)} /></label>
      <label htmlFor="program-mail-body">Message<textarea id="program-mail-body" required minLength={2} maxLength={5000} rows={10} value={draft.body} onChange={event => change('body', event.target.value)} /></label>
      <p className="program-admin-help">Personalize with <code>{'{{name}}'}</code>, <code>{'{{program}}'}</code>, <code>{'{{decision}}'}</code> or <code>{'{{application_url}}'}</code>. Only released decisions appear in emails.</p>
      <button className="app-button app-button-outline" type="submit" disabled={busy}>Preview and review recipients</button>
    </fieldset>
    {preview && <div className="program-mail-confirm"><p><strong>{preview.count}</strong> eligible {preview.count === 1 ? 'applicant' : 'applicants'}</p>{preview.count > 0 ? <label className="program-admin-check"><Checkbox checked={confirmed} disabled={busy} onCheckedChange={setConfirmed} />Send this message to {preview.count} {preview.count === 1 ? 'applicant' : 'applicants'}.</label> : <p>No eligible applicants match this audience.</p>}
      {uncertain && <p role="status">The previous request could not be confirmed. Retry safely before changing the message.</p>}
      <button className="app-button app-button-primary" type="button" disabled={busy || !confirmed || preview.count < 1} onClick={() => void send()}>{busy ? 'Working…' : uncertain ? 'Retry send request safely' : `Send to ${preview.count} ${preview.count === 1 ? 'applicant' : 'applicants'}`}</button></div>}
    {error && <p className="program-admin-error" role="alert">{error.message}</p>}<ProgramAuthRecovery error={error} />{notice && <p className="program-admin-success" role="status">{notice}</p>}
  </form><aside className="program-mail-preview"><h2>Your email preview</h2>{preview?.html ? <><p>Preview for {preview.recipient}</p><strong>{preview.subject}</strong><iframe title="Personalized applicant email preview" sandbox="" srcDoc={preview.html} /></> : <p>Write your message and preview it to review the real recipient count and a personalized email.</p>}</aside></div>;
}
