"use client";
import { useEffect, useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Select } from '@hackways/ui/select';
import { Checkbox } from '@hackways/ui/checkbox';
import { Button } from '@/components/ui/Button';
import { campaignConfiguration } from '@/lib/campaign-design';
import { campaignRequest, type CampaignValues } from '@/lib/campaign-api';
import { CampaignComposer } from './CampaignComposer';

function ServiceMessage() {
  const router = useRouter(), lock = useRef(false), receipt = useRef<{ body: string; key: string } | null>(null);
  const [recipients, setRecipients] = useState(''), [subject, setSubject] = useState(''), [body, setBody] = useState('');
  const [label, setLabel] = useState(''), [url, setURL] = useState(''), [confirmed, setConfirmed] = useState(false);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [preview, setPreview] = useState(''), [previewError, setPreviewError] = useState('');
  const values: CampaignValues = { recipients: recipients.split(/[\s,;]+/).filter(Boolean), subject, body, action_label: label, action_url: url, design: campaignConfiguration().design, confirmed };
  const fingerprint = JSON.stringify(values);
  function change(action: () => void) { action(); setConfirmed(false); setPreview(''); }
  useEffect(() => {
    const input = JSON.parse(fingerprint) as CampaignValues;
    if (!input.subject || !input.body) return;
    const controller = new AbortController();
    const timer = window.setTimeout(async () => {
      try { const result = await campaignRequest('/messages/preview', input, '', undefined, controller.signal); if (!controller.signal.aborted) { setPreview(result.html || ''); setPreviewError(''); } }
      catch (failure) { if (!controller.signal.aborted) setPreviewError(failure instanceof Error ? failure.message : 'Preview is unavailable.'); }
    }, 600);
    return () => { window.clearTimeout(timer); controller.abort(); };
  }, [fingerprint]);
  async function send(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current) return;
    if (receipt.current?.body !== fingerprint) receipt.current = { body: fingerprint, key: crypto.randomUUID() };
    lock.current = true; setBusy(true); setError('');
    try { await campaignRequest('/messages/send', values, '', receipt.current.key); router.push('/messaging'); router.refresh(); }
    catch (failure) { setError(failure instanceof Error ? failure.message : 'The message could not be queued. Your input is still here.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <div className="campaign-studio"><form className="campaign-composer" onSubmit={send} aria-busy={busy}>
    <fieldset disabled={busy} className="campaign-editor-fields">
      <div className="field"><label htmlFor="message-recipients">To</label><textarea id="message-recipients" className="app-input" rows={4} value={recipients} maxLength={13000} placeholder="Email addresses, one per line" required onChange={event => change(() => setRecipients(event.target.value))} /><p className="field-help">Up to 50 verified Hackways users. For news and offers, select Promotional campaign.</p></div>
      <div className="field"><label htmlFor="message-subject">Subject</label><input id="message-subject" className="app-input" value={subject} maxLength={150} required onChange={event => change(() => setSubject(event.target.value))} /></div>
      <div className="field"><label htmlFor="message-body">Message</label><textarea id="message-body" className="app-input" rows={10} value={body} maxLength={20000} required onChange={event => change(() => setBody(event.target.value))} /></div>
      <div className="campaign-fields"><div className="field"><label htmlFor="message-label">Button label (optional)</label><input id="message-label" className="app-input" value={label} maxLength={80} onChange={event => change(() => setLabel(event.target.value))} /></div><div className="field"><label htmlFor="message-url">Button link</label><input id="message-url" className="app-input" type="url" value={url} maxLength={2000} placeholder="https://" onChange={event => change(() => setURL(event.target.value))} /></div></div>
      <p className="campaign-from">From notify@mail.hackways.in · Replies to hello@hackways.in</p>
      <label className="campaign-confirm"><Checkbox checked={confirmed} disabled={busy} onCheckedChange={setConfirmed} required />I reviewed this message and its recipients.</label>
    </fieldset>
    {error && <p role="alert" className="journey-error">{error}</p>}
    <footer className="campaign-editor-footer"><Button type="submit" disabled={busy || !confirmed || !preview}>{busy ? 'Queuing…' : 'Queue message'}</Button><span>Delivery status appears in Messaging</span></footer>
  </form><aside className="campaign-live-preview"><header><h2>Email preview</h2></header><div className="campaign-preview-canvas">{preview ? <iframe title="Custom service message preview" srcDoc={preview} sandbox="" referrerPolicy="no-referrer" /> : <p role="status">{previewError || 'Add a subject and message to preview.'}</p>}</div></aside></div>;
}

export function MessageComposer() {
  const [purpose, setPurpose] = useState<'service' | 'promotion'>('service');
  return <><div className="message-purpose"><label htmlFor="message-purpose">Message type</label><Select id="message-purpose" label="Message type" value={purpose} options={[{ value: 'service', label: 'Service or event update' }, { value: 'promotion', label: 'Promotional campaign' }] as const} onValueChange={setPurpose} /></div>
    <section className="message-purpose-panel" hidden={purpose !== 'service'} aria-label="Service message"><ServiceMessage /></section>
    <section className="message-purpose-panel" hidden={purpose !== 'promotion'} aria-label="Promotional campaign"><CampaignComposer /></section>
  </>;
}
