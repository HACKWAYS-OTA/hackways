"use client";
import { useEffect, useId, useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Select } from '@hackways/ui/select';
import { Button } from '@/components/ui/Button';
import { Icon } from '@/components/ui/Icon';
import type { Campaign } from '@/lib/campaigns';
import { campaignConfiguration, type CampaignBlock, type CampaignDesign, type CampaignConfiguration } from '@/lib/campaign-design';
import { campaignRequest, type CampaignValues } from '@/lib/campaign-api';

function TextField({ label, value, onChange, max = 180, rows, type = 'text', required = false, placeholder }: {
  label: string; value: string; onChange: (value: string) => void; max?: number; rows?: number; type?: string; required?: boolean; placeholder?: string;
}) {
  const id = useId();
  return <div className="field"><label htmlFor={id}>{label}</label>{rows
    ? <textarea id={id} className="app-input" value={value} onChange={event => onChange(event.target.value)} maxLength={max} rows={rows} required={required} placeholder={placeholder} />
    : <input id={id} className="app-input" value={value} onChange={event => onChange(event.target.value)} maxLength={max} type={type} required={required} placeholder={placeholder} />}</div>;
}

const blockOptions = [{ value: 'heading', label: 'Heading' }, { value: 'text', label: 'Paragraph' }, { value: 'image', label: 'Image' }, { value: 'button', label: 'Button' }, { value: 'divider', label: 'Divider' }] as const;
function BlockEditor({ block, index, count, update, move, remove }: {
  block: CampaignBlock; index: number; count: number; update: (value: CampaignBlock) => void; move: (offset: number) => void; remove: () => void;
}) {
  return <section className="campaign-block"><header><Select label={`Block ${index + 1} type`} value={block.kind} options={blockOptions} onValueChange={kind => update({ ...block, kind })} /><div>
    <Button variant="outline" size="sm" disabled={index === 0} onClick={() => move(-1)} aria-label={`Move block ${index + 1} up`}><Icon name="up" size={16} /></Button>
    <Button variant="outline" size="sm" disabled={index === count - 1} onClick={() => move(1)} aria-label={`Move block ${index + 1} down`}><Icon name="down" size={16} /></Button>
    <Button variant="ghost" size="sm" onClick={remove} aria-label={`Remove block ${index + 1}`}><Icon name="close" size={16} /></Button>
  </div></header>
    {block.kind !== 'divider' && block.kind !== 'image' && <TextField label={block.kind === 'button' ? 'Button label' : block.kind === 'heading' ? 'Heading' : 'Paragraph'} value={block.text} rows={block.kind === 'text' ? 4 : undefined} max={block.kind === 'button' ? 80 : block.kind === 'heading' ? 180 : 4000} onChange={text => update({ ...block, text })} />}
    {(block.kind === 'image' || block.kind === 'button') && <TextField label={block.kind === 'image' ? 'Image URL' : 'Link URL'} value={block.url} type="url" max={2000} placeholder="https://" onChange={url => update({ ...block, url })} />}
    {block.kind === 'image' && <TextField label="Image description" value={block.alt} onChange={alt => update({ ...block, alt })} />}
  </section>;
}

export function CampaignComposer({ campaign, onDirtyChange }: { campaign?: Campaign; onDirtyChange?: () => void }) {
  const router = useRouter(), lock = useRef(false), receipt = useRef<{ body: string; key: string } | null>(null);
  const [title, setTitle] = useState(campaign?.title || ''), [subject, setSubject] = useState(campaign?.subject || ''), [body, setBody] = useState(campaign?.body || '');
  const [actionLabel, setActionLabel] = useState(campaign?.action_label || ''), [actionURL, setActionURL] = useState(campaign?.action_url || '');
  const [configuration, setConfiguration] = useState(() => campaignConfiguration(campaign?.configuration));
  const [recipients, setRecipients] = useState(configuration.recipients.join('\n'));
  const [section, setSection] = useState('content'), [busy, setBusy] = useState(false), [error, setError] = useState('');
  const [preview, setPreview] = useState(''), [previewError, setPreviewError] = useState(''), [previewBusy, setPreviewBusy] = useState(false), [count, setCount] = useState<number | null>(null);
  const [previewWidth, setPreviewWidth] = useState<'desktop' | 'mobile'>('desktop');
  const design = configuration.design;
  function change(action: () => void) { action(); setPreview(''); setPreviewError(''); onDirtyChange?.(); }
  function config(patch: Partial<CampaignConfiguration>) { change(() => setConfiguration(current => ({ ...current, ...patch }))); }
  function style(patch: Partial<CampaignDesign>) { change(() => setConfiguration(current => ({ ...current, design: { ...current.design, ...patch } }))); }
  function blocks(value: CampaignBlock[]) { style({ blocks: value }); }
  const values: CampaignValues = { title, subject, body, action_label: actionLabel, action_url: actionURL, configuration: { ...configuration, recipients: recipients.split(/[\s,;]+/).map(value => value.trim()).filter(Boolean) }, ...(campaign ? { revision: campaign.revision } : {}) };
  const fingerprint = JSON.stringify(values);
  useEffect(() => {
    const input = JSON.parse(fingerprint) as CampaignValues;
    if (!input.subject || !input.body) { setPreviewBusy(false); return; }
    const controller = new AbortController(); setPreviewBusy(true);
    const timeout = window.setTimeout(async () => {
      try {
        const result = await campaignRequest('/campaigns/preview', { ...input, title: input.title || 'Email preview' }, '', undefined, controller.signal);
        if (!controller.signal.aborted) { setPreview(result.html || ''); setCount(result.audience_count ?? null); setPreviewError(''); }
      } catch (failure) { if (!controller.signal.aborted) setPreviewError(failure instanceof Error ? failure.message : 'Preview is unavailable.'); }
      finally { if (!controller.signal.aborted) setPreviewBusy(false); }
    }, 600);
    return () => { window.clearTimeout(timeout); controller.abort(); };
  }, [fingerprint]);
  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current) return;
    if (receipt.current?.body !== fingerprint) receipt.current = { body: fingerprint, key: crypto.randomUUID() };
    lock.current = true; setBusy(true); setError('');
    try {
      const result = await campaignRequest(campaign ? `/campaigns/${campaign.reference}/edit` : '/campaigns', values, '', receipt.current.key);
      if (campaign) router.refresh(); else if (result.data) { router.replace(`/campaigns/details/${result.data.reference}`); router.refresh(); }
    } catch (failure) { setError(failure instanceof Error ? failure.message : 'The draft could not be saved. Your changes are still here.'); }
    finally { lock.current = false; setBusy(false); }
  }
  const steps = [['content', 'Content'], ['design', 'Design'], ['recipients', 'Recipients'], ['delivery', 'Delivery']] as const;
  return <div className="campaign-studio"><form onSubmit={save} className="campaign-composer" aria-busy={busy} noValidate>
    <nav className="campaign-editor-nav" aria-label="Campaign editor">{steps.map(([key, label]) => <button type="button" key={key} aria-current={section === key ? 'step' : undefined} onClick={() => setSection(key)}>{label}</button>)}</nav>
    <fieldset disabled={busy}>
      {section === 'content' && <div className="campaign-editor-fields">
        <TextField label="Campaign name" value={title} max={150} onChange={value => change(() => setTitle(value))} required />
        <TextField label="Email subject" value={subject} max={150} onChange={value => change(() => setSubject(value))} required />
        <TextField label="Inbox preview text" value={design.preheader} max={200} onChange={preheader => style({ preheader })} />
        <TextField label="Email heading" value={design.heading} placeholder="Use the email subject" onChange={heading => style({ heading })} />
        <TextField label="Message" value={body} max={20000} rows={7} onChange={value => change(() => setBody(value))} required />
        <div className="campaign-blocks">{design.blocks.map((block, index) => <BlockEditor key={index} block={block} index={index} count={design.blocks.length}
          update={value => blocks(design.blocks.map((item, position) => position === index ? value : item))}
          move={offset => { const next = [...design.blocks]; [next[index], next[index + offset]] = [next[index + offset], next[index]]; blocks(next); }}
          remove={() => blocks(design.blocks.filter((_, position) => position !== index))} />)}</div>
        <Button variant="outline" disabled={design.blocks.length >= 12} onClick={() => blocks([...design.blocks, { kind: 'text', text: '', url: '', alt: '' }])}>Add content block</Button>
        <div className="campaign-fields"><TextField label="Main button label" value={actionLabel} max={80} onChange={value => change(() => setActionLabel(value))} /><TextField label="Main button link" value={actionURL} type="url" max={2000} placeholder="https://" onChange={value => change(() => setActionURL(value))} /></div>
      </div>}
      {section === 'design' && <div className="campaign-editor-fields">
        <TextField label="Brand name" value={design.brand} max={80} onChange={brand => style({ brand })} />
        <TextField label="Banner image URL" value={design.banner_url} type="url" max={2000} placeholder="https://" onChange={banner_url => style({ banner_url })} />
        {design.banner_url && <TextField label="Banner description" value={design.banner_alt} onChange={banner_alt => style({ banner_alt })} />}
        <div className="campaign-fields"><label className="field">Layout<Select label="Email layout" value={design.layout} options={[{ value: 'card', label: 'Inset panel' }, { value: 'plain', label: 'Plain' }] as const} onValueChange={layout => style({ layout })} /></label>
          <label className="field">Alignment<Select label="Email alignment" value={design.align} options={[{ value: 'left', label: 'Left' }, { value: 'center', label: 'Center' }] as const} onValueChange={align => style({ align })} /></label></div>
        <label className="field">Font<Select label="Email font" value={design.font} options={[{ value: 'sans', label: 'Arial · Sans serif' }, { value: 'serif', label: 'Georgia · Serif' }, { value: 'mono', label: 'Courier · Monospace' }] as const} onValueChange={font => style({ font })} /></label>
        <div className="campaign-color-grid">{([['background', 'Page'], ['surface', 'Panel'], ['text', 'Text'], ['accent', 'Buttons'], ['button_text', 'Button text']] as const).map(([key, label]) => <label className="campaign-color" key={key}><span>{label}</span><input type="color" value={design[key]} aria-label={`${label} color`} onChange={event => style({ [key]: event.target.value })} /><input className="app-input" value={design[key]} aria-label={`${label} hex color`} maxLength={7} onChange={event => style({ [key]: event.target.value })} /></label>)}</div>
      </div>}
      {section === 'recipients' && <div className="campaign-editor-fields">
        <label className="field">Audience<Select label="Campaign audience" value={configuration.audience} options={[{ value: 'all', label: 'All subscribed contacts' }, { value: 'account', label: 'Verified account subscribers' }, { value: 'import', label: 'Imported opt-in contacts' }] as const} onValueChange={audience => config({ audience })} /></label>
        <TextField label="Specific email addresses (optional)" value={recipients} rows={8} max={52000} placeholder="One email per line" onChange={value => change(() => setRecipients(value))} />
        <p className="field-help">Restrict the selected audience to these addresses. Up to 200; only currently subscribed contacts qualify. Leave empty to use the whole selected audience.</p>
        {count !== null && <p className="campaign-recipient-count">{count} eligible recipients</p>}
        <a className="app-button app-button-outline" href="/campaigns/contacts">Manage contacts</a>
      </div>}
      {section === 'delivery' && <div className="campaign-editor-fields">
        <label className="field">Send timing<Select label="Send timing" value={configuration.scheduled_at ? 'scheduled' : 'now'} options={[{ value: 'now', label: 'Send when launched' }, { value: 'scheduled', label: 'Schedule a date and time' }] as const} onValueChange={mode => config({ scheduled_at: mode === 'now' ? null : new Date(Date.now() + 3600000).toISOString() })} /></label>
        {configuration.scheduled_at && <label className="field">Send at ({Intl.DateTimeFormat().resolvedOptions().timeZone})<input className="app-input" type="datetime-local" value={new Date(new Date(configuration.scheduled_at).getTime() - new Date(configuration.scheduled_at).getTimezoneOffset() * 60000).toISOString().slice(0, 16)} onChange={event => { const date = new Date(event.target.value); if (!Number.isNaN(date.getTime())) config({ scheduled_at: date.toISOString() }); }} /></label>}
        <p className="field-help">Saving keeps a draft. Launch from the saved campaign after reviewing recipients and sending a test.</p>
        <dl className="campaign-sender"><div><dt>From</dt><dd>updates@mail.hackways.in</dd></div><div><dt>Replies</dt><dd>hello@hackways.in</dd></div></dl>
      </div>}
    </fieldset>
    {error && <p className="journey-error" role="alert">{error}</p>}
    <footer className="campaign-editor-footer"><Button type="submit" disabled={busy}>{busy ? 'Saving…' : 'Save draft'}</Button><span>Unsubscribe link included automatically</span></footer>
  </form>
  <aside className="campaign-live-preview"><header><h2>Email preview</h2><Select label="Preview width" value={previewWidth} options={[{ value: 'desktop', label: 'Desktop' }, { value: 'mobile', label: 'Mobile' }] as const} onValueChange={setPreviewWidth} /></header>
    <div className={`campaign-preview-canvas is-${previewWidth}`} aria-busy={previewBusy}>{preview ? <iframe title="Current campaign email preview" srcDoc={preview} sandbox="" referrerPolicy="no-referrer" /> : <p role="status">{previewBusy ? 'Updating preview…' : previewError || 'Add a subject and message to preview your email.'}</p>}</div>
    {previewError && preview && <p role="alert">{previewError}</p>}
  </aside></div>;
}
