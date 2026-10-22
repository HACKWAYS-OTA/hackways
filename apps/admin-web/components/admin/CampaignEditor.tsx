"use client";
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Checkbox } from '@hackways/ui/checkbox';
import { Button } from '@/components/ui/Button';
import type { Campaign } from '@/lib/campaigns';

type Values = Record<string, string | number | boolean>;
async function mutate(path: string, values: Values, reason: string, key?: string) {
  const base = (process.env.NEXT_PUBLIC_API_URL || '/api/v1').replace(/\/$/, '');
  const controller = new AbortController(), timeout = window.setTimeout(() => controller.abort(), 15000);
  try {
    const response = await fetch(`${base}/admin/global${path}`, { method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json', 'X-Admin-Reason': encodeURIComponent(reason), ...(key ? { 'Idempotency-Key': key } : {}) }, body: JSON.stringify(values), signal: controller.signal });
    if (response.status === 401 || response.status === 403) { window.location.assign(response.status === 401 ? '/login?next=%2Fcampaigns' : '/access-denied'); throw new Error('Admin sign-in is required.'); }
    const result = await response.json() as { data?: Campaign; message?: string; error?: { message?: string }; queued?: number; imported?: number; unchanged?: number };
    if (!response.ok) throw new Error(result.error?.message || 'The action could not be saved. Your input is still here.');
    return result;
  } finally { window.clearTimeout(timeout); }
}
export function CampaignEditor({ campaign }: { campaign?: Campaign }) {
  const router = useRouter(), lock = useRef(false), receipt = useRef<{ body: string; key: string } | null>(null);
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current) return;
    const form = new FormData(event.currentTarget), values: Values = { title: String(form.get('title')), subject: String(form.get('subject')), body: String(form.get('body')), action_label: String(form.get('action_label')), action_url: String(form.get('action_url')), ...(campaign ? { revision: campaign.revision } : {}) };
    const canonical = JSON.stringify(values);
    if (receipt.current?.body !== canonical) receipt.current = { body: canonical, key: crypto.randomUUID() };
    lock.current = true; setBusy(true); setError('');
    try { const result = await mutate(campaign ? `/campaigns/${campaign.reference}/edit` : '/campaigns', values, String(form.get('reason')), receipt.current.key); if (campaign) router.refresh(); else if (result.data) { router.replace(`/campaigns/details/${result.data.reference}`); router.refresh(); } }
    catch (err) { setError(err instanceof Error ? err.message : 'Could not save the draft.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <form className="campaign-form" onSubmit={submit} aria-busy={busy}>
    <div className="field"><label htmlFor="campaign-title">Campaign name</label><input id="campaign-title" name="title" className="app-input" defaultValue={campaign?.title} maxLength={150} required disabled={busy} /></div>
    <div className="field"><label htmlFor="campaign-subject">Email subject</label><input id="campaign-subject" name="subject" className="app-input" defaultValue={campaign?.subject} maxLength={150} required disabled={busy} /></div>
    <div className="field"><label htmlFor="campaign-body">Message</label><textarea id="campaign-body" name="body" className="app-input" rows={10} defaultValue={campaign?.body} maxLength={20000} required disabled={busy} /><p className="field-help">Plain text and paragraphs. The email layout and unsubscribe link are added automatically.</p></div>
    <div className="campaign-fields"><div className="field"><label htmlFor="campaign-action">Button label (optional)</label><input id="campaign-action" name="action_label" className="app-input" defaultValue={campaign?.action_label} maxLength={80} disabled={busy} /></div><div className="field"><label htmlFor="campaign-url">Button URL (HTTPS)</label><input id="campaign-url" name="action_url" type="url" className="app-input" defaultValue={campaign?.action_url} maxLength={2000} disabled={busy} /></div></div>
    <div className="field"><label htmlFor="campaign-reason">Reason for this change</label><input id="campaign-reason" name="reason" className="app-input" minLength={3} maxLength={1000} required disabled={busy} /></div>
    {error && <p className="journey-error" role="alert">{error}</p>}<Button type="submit" disabled={busy}>{busy ? 'Saving...' : 'Save draft'}</Button>
  </form>;
}
export function CampaignActions({ campaign, audienceCount }: { campaign: Campaign; audienceCount: number }) {
  const router = useRouter(), lock = useRef(false);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [notice, setNotice] = useState('');
  async function act(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current) return;
    const form = new FormData(event.currentTarget), action = String(form.get('action'));
    lock.current = true; setBusy(action); setError(''); setNotice('');
    try { const result = await mutate(`/campaigns/${campaign.reference}/${action}`, { revision: campaign.revision, audience_count: audienceCount, confirmed: form.get('confirmed') === 'yes' }, String(form.get('reason'))); setNotice(action === 'test' ? result.queued ? 'Test email queued for your verified admin address.' : 'A test is already queued for this draft version.' : action === 'cancel' ? 'Campaign cancelled. Provider-accepted email cannot be recalled.' : 'Campaign queued. Track provider acceptance below.'); router.refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : 'Campaign action failed.'); }
    finally { lock.current = false; setBusy(''); }
  }
  if (campaign.status === 'cancelled') return <p role="status">This campaign is cancelled.</p>;
  return <div className="campaign-controls">
    {campaign.status === 'draft' && <form onSubmit={act}><h2>Test email</h2><p>Send this saved version once to your verified admin email. No audience is contacted.</p><input type="hidden" name="action" value="test" /><label className="field">Reason<input className="app-input" name="reason" minLength={3} maxLength={1000} required disabled={!!busy} /></label><Button variant="outline" type="submit" disabled={!!busy}>{busy === 'test' ? 'Queuing...' : 'Send test to me'}</Button></form>}
    {campaign.status === 'draft' && <form onSubmit={act}><h2>Send campaign</h2><p><strong>{audienceCount}</strong> eligible recipients. From updates@mail.hackways.in.</p><input type="hidden" name="action" value="send" /><label className="campaign-confirm"><Checkbox name="confirmed" value="yes" required disabled={!!busy} />I reviewed this saved email and its opted-in audience.</label><label className="field">Reason<input className="app-input" name="reason" minLength={3} maxLength={1000} required disabled={!!busy} /></label><Button type="submit" disabled={!!busy || audienceCount < 1 || audienceCount > 1000}>{busy === 'send' ? 'Queuing...' : `Send to ${audienceCount} recipients`}</Button>{audienceCount > 1000 && <p role="alert">This audience exceeds the 1,000-recipient campaign limit.</p>}</form>}
    <form onSubmit={act}><h2>Cancel campaign</h2><p>Stop unsent messages. Email already accepted by the provider cannot be recalled.</p><input type="hidden" name="action" value="cancel" /><label className="field">Reason<input className="app-input" name="reason" minLength={3} maxLength={1000} required disabled={!!busy} /></label><Button type="submit" variant="outline" disabled={!!busy}>{busy === 'cancel' ? 'Cancelling...' : 'Cancel campaign'}</Button></form>
    {error && <p className="journey-error" role="alert">{error}</p>}{notice && <p className="journey-notice" role="status">{notice}</p>}
  </div>;
}
export function PromotionImport() {
  const router = useRouter(), lock = useRef(false);
  const [csv, setCSV] = useState(''), [busy, setBusy] = useState(false), [error, setError] = useState(''), [notice, setNotice] = useState('');
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current) return;
    const form = new FormData(event.currentTarget), date = new Date(String(form.get('consented_at')));
    if (Number.isNaN(date.getTime())) { setError('Enter the consent date.'); return; }
    lock.current = true; setBusy(true); setError(''); setNotice('');
    try { const result = await mutate('/campaign-contacts/import', { csv, consent_source: String(form.get('consent_source')), consented_at: date.toISOString(), confirmed: form.get('confirmed') === 'yes' }, String(form.get('reason'))); setNotice(`${result.imported} contacts imported; ${result.unchanged} existing contacts preserved.`); router.refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : 'Import failed. No contacts were changed.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <form className="campaign-form" onSubmit={submit}><h2>Import opted-in contacts</h2><p>Up to 200 contacts per import. Existing contacts and unsubscribes are preserved.</p><div className="field"><label htmlFor="contact-file">CSV file (email,name)</label><input id="contact-file" type="file" accept=".csv,text/csv" disabled={busy} onChange={async event => { const file = event.target.files?.[0]; if (!file) return; if (file.size > 65536) { setError('Choose a CSV smaller than 64 KB.'); return; } try { setCSV(await file.text()); setError(''); } catch { setError('The file could not be read.'); } }} /></div><div className="field"><label htmlFor="contacts-csv">Review CSV</label><textarea id="contacts-csv" className="app-input" rows={6} value={csv} onChange={event => setCSV(event.target.value)} maxLength={65536} required disabled={busy} /></div><label className="field">Consent source or evidence<input className="app-input" name="consent_source" maxLength={1000} required disabled={busy} /></label><label className="field">Consent date<input className="app-input" name="consented_at" type="datetime-local" required disabled={busy} /></label><label className="campaign-confirm"><Checkbox name="confirmed" value="yes" required disabled={busy} />Every contact explicitly opted in to Hackways promotional email.</label><label className="field">Reason for import<input className="app-input" name="reason" minLength={3} maxLength={1000} required disabled={busy} /></label>{error && <p className="journey-error" role="alert">{error}</p>}{notice && <p className="journey-notice" role="status">{notice}</p>}<Button type="submit" disabled={busy}>{busy ? 'Importing...' : 'Import contacts'}</Button></form>;
}
