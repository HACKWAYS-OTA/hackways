'use client';

import { useEffect, useId, useState } from 'react';
import { Select } from '@hackways/ui/select';
import { Checkbox } from '@hackways/ui/checkbox';
import { Button } from '@/components/ui/Button';
import { adminRequest } from '@/lib/admin-request';
import type { CampaignConfiguration } from '@/lib/campaign-design';

type Recipient = { reference: number; email: string; display_name: string; source: string };

export function CampaignAudience({ configuration, recipients, setRecipients, update, count }: {
  configuration: CampaignConfiguration; recipients: string; setRecipients: (value: string) => void;
  update: (value: Partial<CampaignConfiguration>) => void; count: number | null;
}) {
  const id = useId();
  const [query, setQuery] = useState(''), [results, setResults] = useState<Recipient[]>([]);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [more, setMore] = useState(false);
  const selected = Array.from(new Set(recipients.split(/[\s,;]+/).map(value => value.trim().toLowerCase()).filter(Boolean)));

  useEffect(() => {
    if (configuration.recipient_mode !== 'selected') return;
    const controller = new AbortController(); setBusy(true); setError(''); setResults([]);
    const timer = window.setTimeout(async () => {
      try {
        const values = new URLSearchParams({ q: query.trim(), source: configuration.audience });
        const response = await adminRequest<{ data: Recipient[]; more: boolean }>(`global/campaigns/recipients?${values}`, { signal: controller.signal });
        if (!controller.signal.aborted) { setResults(response.data); setMore(response.more); }
      } catch (failure) {
        if (!controller.signal.aborted) setError(failure instanceof Error ? failure.message : 'Recipients could not be loaded. Try another search.');
      } finally { if (!controller.signal.aborted) setBusy(false); }
    }, 300);
    return () => { window.clearTimeout(timer); controller.abort(); };
  }, [query, configuration.audience, configuration.recipient_mode]);

  function choose(email: string, checked: boolean) {
    setRecipients((checked ? Array.from(new Set([...selected, email])) : selected.filter(value => value !== email)).join('\n'));
  }

  return <div className="campaign-editor-fields">
    <Select label="Audience" value={configuration.audience} options={[
      { value: 'account', label: 'All verified users with promotional email enabled' },
      { value: 'import', label: 'Imported opt-in contacts' },
      { value: 'all', label: 'All subscribed users and contacts' },
    ] as const} onValueChange={audience => update({ audience })} />
    <Select label="Recipients" value={configuration.recipient_mode} options={[
      { value: 'all', label: 'Everyone in this audience' },
      { value: 'selected', label: 'Specific people' },
    ] as const} onValueChange={recipient_mode => update({ recipient_mode })} />
    <p className="field-help">Unsubscribed addresses and unverified accounts are excluded at send time.</p>
    {configuration.recipient_mode === 'selected' && <>
      <div className="field"><label htmlFor={`${id}-search`}>Find people</label><input id={`${id}-search`} className="app-input" type="search" value={query} maxLength={200} placeholder="Start a name or email address" onChange={event => setQuery(event.target.value)} /></div>
      <div className="campaign-recipient-results" aria-busy={busy}>
        {busy ? <p role="status">Finding recipients...</p> : error ? <p role="alert">{error}</p> : results.length ? results.map(person => <label key={person.reference} className="campaign-recipient-option">
          <Checkbox checked={selected.includes(person.email)} disabled={selected.length >= 200 && !selected.includes(person.email)} onCheckedChange={checked => choose(person.email, checked)} />
          <span><strong>{person.display_name || person.email}</strong>{person.display_name && <small>{person.email}</small>}</span><small>{person.source === 'import' ? 'Imported' : 'Account'}</small>
        </label>) : <p role="status">No eligible recipients match this search.</p>}
        {!busy && more && <p className="field-help">Showing 20 results. Refine the search to find more people.</p>}
      </div>
      {selected.length > 0 && <div className="campaign-selected-recipients" aria-label="Selected recipients">{selected.map(email => <span key={email}>{email}<Button variant="ghost" size="sm" aria-label={`Remove ${email}`} onClick={() => choose(email, false)}>Remove</Button></span>)}</div>}
      <details className="campaign-address-entry"><summary>Paste email addresses</summary><div className="field"><label htmlFor={`${id}-emails`}>Specific addresses</label><textarea id={`${id}-emails`} className="app-input" rows={4} value={recipients} maxLength={52000} onChange={event => setRecipients(event.target.value)} /><p className="field-help">Up to 200 addresses, separated by lines or commas. Only eligible subscribers will receive the campaign.</p></div></details>
      <p className="field-help">{selected.length} selected</p>
    </>}
    {count !== null && <p className="campaign-recipient-count" role="status">{count} eligible recipients</p>}
    <a className="app-button app-button-outline" href="/campaigns/contacts">Manage contacts</a>
  </div>;
}
