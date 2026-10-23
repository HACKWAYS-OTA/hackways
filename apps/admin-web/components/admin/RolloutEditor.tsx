'use client';
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Checkbox } from '@hackways/ui/checkbox';
import { RadioGroup } from '@hackways/ui/radio-group';
import { adminRequest, mutationHeaders } from '@/lib/admin-request';
import type { FeatureRule, RolloutData } from '@/lib/rollouts';

const local = (value?: string) => { if (!value) return ''; const date = new Date(value); return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16); };
const utc = (value: string) => value ? new Date(value).toISOString() : undefined;
export function RolloutEditor({ initial }: { initial: RolloutData }) {
  const router = useRouter();
  const [flags, setFlags] = useState(initial.data.flags);
  const [label, setLabel] = useState('');
  const [restore, setRestore] = useState<number | null>(null);
  const [confirmed, setConfirmed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const receipt = useRef<{ signature: string; key: string }>();
  const update = (key: string, patch: Partial<FeatureRule>) => { setMessage(''); setFlags(value => ({ ...value, [key]: { ...value[key], ...patch } })); };
  async function submit(event: FormEvent) {
    event.preventDefault(); if (busy || !confirmed) return;
    const body = JSON.stringify({ revision: initial.data.revision, label, flags: restore ? undefined : flags, restore_version: restore || 0 });
    const signature = body;
    if (receipt.current?.signature !== signature) receipt.current = { signature, key: crypto.randomUUID() };
    setBusy(true); setError(''); setMessage('');
    try { await adminRequest('global/rollouts', { method: 'POST', body, headers: mutationHeaders(receipt.current.key) }); setMessage('Configuration published. Open pages adopt it on their next navigation or scheduled revalidation.'); router.refresh(); }
    catch (failure) { setError(failure instanceof Error ? failure.message : 'Configuration could not be published.'); }
    finally { setBusy(false); }
  }
  const source = restore ? initial.history.find(row => row.reference === restore) : null;
  return <div className="rollout-workbench"><form onSubmit={submit} className="rollout-form"><fieldset disabled={busy}>
    <div className="admin-section-heading"><h2>Feature flags</h2><button type="button" className="app-button app-button-outline" onClick={() => { setFlags(Object.fromEntries(initial.definitions.map(item => [item.key, { ...flags[item.key], enabled: false, percentage: flags[item.key]?.percentage || 0 }]))); setRestore(null); }}>Disable all in draft</button></div>
    <p className="admin-secondary-copy">Permissions and data validation always remain active. Only implemented features appear here.</p>
    {source ? <section className="rollout-restore"><h3>Restore {source.label}</h3><p>Creates a new version using configuration #{source.reference}. Original dates are preserved.</p><button type="button" className="app-button app-button-outline" onClick={() => setRestore(null)}>Cancel restore</button></section> : initial.definitions.map(definition => { const rule = flags[definition.key] || { enabled: false, percentage: 0 }; return <section key={definition.key} className="rollout-feature"><div className="admin-section-heading"><h3>{definition.name}</h3><label className="campaign-confirm"><Checkbox checked={rule.enabled} onCheckedChange={enabled => update(definition.key, { enabled })} disabled={busy} />Enabled</label></div><p className="admin-secondary-copy">{definition.surface}</p><label className="field">Eligible users (%)<input className="app-input" type="number" required min={0} max={100} step={0.01} value={rule.percentage} onChange={event => update(definition.key, { percentage: Number(event.target.value) })} /></label><RadioGroup label={`${definition.name} rollout percentage`} value={String(rule.percentage)} disabled={busy} onValueChange={value => update(definition.key, { percentage: Number(value) })} options={[0, 1, 5, 10, 25, 50, 100].map(value => ({ value: String(value), label: `${value}%` }))} /><div className="campaign-fields"><label className="field">Starts (optional)<input type="datetime-local" className="app-input" value={local(rule.starts_at)} onChange={event => update(definition.key, { starts_at: utc(event.target.value) })} /></label><label className="field">Ends (optional)<input type="datetime-local" className="app-input" value={local(rule.ends_at)} onChange={event => update(definition.key, { ends_at: utc(event.target.value) })} /></label></div></section>; })}
    <label className="field">Configuration version name<input className="app-input" required maxLength={80} value={label} onChange={event => setLabel(event.target.value)} placeholder="For example, Overview rollout 10%" /></label><label className="campaign-confirm"><Checkbox required checked={confirmed} onCheckedChange={setConfirmed} disabled={busy} />I reviewed the percentages, dates and fallback behavior.</label>{error && <p className="field-error" role="alert">{error}</p>}{message && <p className="admin-action-success" role="status">{message}</p>}<button className="app-button app-button-primary" disabled={busy || !confirmed}>{busy ? 'Publishing…' : restore ? 'Publish restored configuration' : 'Publish configuration'}</button>
  </fieldset></form><section className="rollout-history" aria-labelledby="configuration-history"><h2 id="configuration-history">Version history</h2><p className="admin-secondary-copy">Active: {initial.data.label} · #{initial.data.version}</p><ol>{initial.history.map(version => <li key={version.reference}><div><strong>{version.label}</strong><span>#{version.reference} · {new Date(version.created_at).toLocaleString()}</span><small>{version.created_by}{version.restored_from ? ` · Restored from #${version.restored_from}` : ''}</small></div><button type="button" className="app-button app-button-outline app-button-sm" disabled={busy || version.reference === initial.data.version} onClick={() => { setRestore(version.reference); setLabel(`Restore ${version.label}`.slice(0, 80)); setConfirmed(false); }}>Restore</button></li>)}</ol></section></div>;
}
