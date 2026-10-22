"use client";
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Checkbox } from '@hackways/ui/checkbox';
import { Button } from '@/components/ui/Button';
import type { Campaign } from '@/lib/campaigns';

import { campaignRequest as mutate } from '@/lib/campaign-api';
export { CampaignComposer as CampaignEditor } from './CampaignComposer';
export function CampaignActions({ campaign, audienceCount, disabled = false }: { campaign: Campaign; audienceCount: number; disabled?: boolean }) {
  const router = useRouter(), lock = useRef(false);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [notice, setNotice] = useState('');
  async function act(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current || disabled) return;
    const form = new FormData(event.currentTarget), action = String(form.get('action'));
    lock.current = true; setBusy(action); setError(''); setNotice('');
    try { const result = await mutate(`/campaigns/${campaign.reference}/${action}`, { revision: campaign.revision, audience_count: audienceCount, confirmed: form.get('confirmed') === 'yes' }, ''); setNotice(action === 'test' ? result.queued ? 'Test email queued for your verified admin address.' : 'A test is already queued for this draft version.' : action === 'cancel' ? 'Campaign cancelled. Provider-accepted email cannot be recalled.' : campaign.configuration?.scheduled_at ? 'Campaign scheduled. Delivery starts at the saved send time.' : 'Campaign queued. Track provider acceptance below.'); router.refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : 'Campaign action failed.'); }
    finally { lock.current = false; setBusy(''); }
  }
  if (campaign.status === 'cancelled') return <p role="status">This campaign is cancelled.</p>;
  return <div className="campaign-controls">{disabled && <p className="campaign-save-notice" role="status">Save your changes before testing or launching this campaign.</p>}
    {campaign.status === 'draft' && <form onSubmit={act}><h2>Test email</h2><p>Send this saved version once to your verified admin email. No audience is contacted.</p><input type="hidden" name="action" value="test" /><Button variant="outline" type="submit" disabled={!!busy || disabled}>{busy === 'test' ? 'Queuing...' : 'Send test to me'}</Button></form>}
    {campaign.status === 'draft' && <form onSubmit={act}><h2>Send campaign</h2><p><strong>{audienceCount}</strong> eligible recipients. From updates@mail.hackways.in.</p><input type="hidden" name="action" value="send" /><label className="campaign-confirm"><Checkbox name="confirmed" value="yes" required disabled={!!busy || disabled} />I reviewed this saved email and its opted-in audience.</label><Button type="submit" disabled={!!busy || disabled || audienceCount < 1 || audienceCount > 1000}>{busy === 'send' ? 'Queuing...' : `${campaign.configuration?.scheduled_at ? 'Schedule for' : 'Send to'} ${audienceCount} recipients`}</Button>{audienceCount > 1000 && <p role="alert">This audience exceeds the 1,000-recipient campaign limit.</p>}</form>}
    <form onSubmit={act}><h2>Cancel campaign</h2><p>Stop unsent messages. Email already accepted by the provider cannot be recalled.</p><input type="hidden" name="action" value="cancel" /><Button type="submit" variant="outline" disabled={!!busy || disabled}>{busy === 'cancel' ? 'Cancelling...' : 'Cancel campaign'}</Button></form>
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
    try { const result = await mutate('/campaign-contacts/import', { csv, consent_source: String(form.get('consent_source')), consented_at: date.toISOString(), confirmed: form.get('confirmed') === 'yes' }, ''); setNotice(`${result.imported} contacts imported; ${result.unchanged} existing contacts preserved.`); router.refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : 'Import failed. No contacts were changed.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <form className="campaign-form" onSubmit={submit}><h2>Import opted-in contacts</h2><p>Up to 200 contacts per import. Existing contacts and unsubscribes are preserved.</p><div className="field"><label htmlFor="contact-file">CSV file (email,name)</label><input id="contact-file" type="file" accept=".csv,text/csv" disabled={busy} onChange={async event => { const file = event.target.files?.[0]; if (!file) return; if (file.size > 65536) { setError('Choose a CSV smaller than 64 KB.'); return; } try { setCSV(await file.text()); setError(''); } catch { setError('The file could not be read.'); } }} /></div><div className="field"><label htmlFor="contacts-csv">Review CSV</label><textarea id="contacts-csv" className="app-input" rows={6} value={csv} onChange={event => setCSV(event.target.value)} maxLength={65536} required disabled={busy} /></div><label className="field">Consent source or evidence<input className="app-input" name="consent_source" maxLength={1000} required disabled={busy} /></label><label className="field">Consent date<input className="app-input" name="consented_at" type="datetime-local" required disabled={busy} /></label><label className="campaign-confirm"><Checkbox name="confirmed" value="yes" required disabled={busy} />Every contact explicitly opted in to Hackways promotional email.</label>{error && <p className="journey-error" role="alert">{error}</p>}{notice && <p className="journey-notice" role="status">{notice}</p>}<Button type="submit" disabled={busy}>{busy ? 'Importing...' : 'Import contacts'}</Button></form>;
}
