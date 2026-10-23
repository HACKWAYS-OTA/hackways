'use client';
import { useRef, useState, type FormEvent } from 'react';
import { RadioGroup } from '@hackways/ui/radio-group';
import { Checkbox } from '@hackways/ui/checkbox';
import { adminModules, type AdminModule, type AdminRecord } from '@/lib/admin-modules';
import { adminRequest, mutationHeaders } from '@/lib/admin-request';

const actions: Partial<Record<AdminModule, { value: string; label: string }[]>> = {
  events: [{ value: 'published', label: 'Publish' }, { value: 'unpublished', label: 'Unpublish' }, { value: 'cancelled', label: 'Cancel events' }],
  moderation: [{ value: 'published', label: 'Publish' }, { value: 'cancelled', label: 'Cancel events' }],
  spaces: [{ value: 'archived', label: 'Archive' }, { value: 'active', label: 'Restore' }],
  registrations: [{ value: 'confirmed', label: 'Approve' }, { value: 'pending_approval', label: 'Awaiting approval' }, { value: 'waitlisted', label: 'Waiting list' }, { value: 'rejected', label: 'Decline' }, { value: 'cancelled', label: 'Cancel registrations' }],
  tickets: [{ value: 'check-in', label: 'Check in' }], messaging: [{ value: 'retry', label: 'Retry eligible failed emails' }],
};
export const supportsBulkActions = (resource: AdminModule) => Boolean(actions[resource]);
type Result = { reference: number; label: string; message: string; success: boolean };
export function BulkActions({ resource, records, onFinished, onClear, onBusyChange }: { resource: AdminModule; records: AdminRecord[]; onFinished: () => void; onClear: () => void; onBusyChange: (busy: boolean) => void }) {
  const options = actions[resource] || [];
  const [action, setAction] = useState(options[0]?.value || '');
  const [confirmed, setConfirmed] = useState(false), [notify, setNotify] = useState(true);
  const [busy, setBusy] = useState(false), [results, setResults] = useState<Result[]>([]), [progress, setProgress] = useState('');
  const receipts = useRef(new Map<string, { path: string; body: string; key: string }>());
  async function submit(event: FormEvent) {
    event.preventDefault(); if (busy || !confirmed || !records.length || records.length > 20) return;
    const targets = [...records]; setBusy(true); onBusyChange(true); setResults([]);
    const next: Result[] = [];
    for (let index = 0; index < targets.length; index++) {
      const row = targets[index];
      const label = String(row.title || row.name || row.public_code || row.recipient_email || row.reference);
      setProgress(`Processing ${index + 1} of ${targets.length}`);
      try {
        const receiptSignature = JSON.stringify([resource, row.reference, action, notify]);
        let receipt = receipts.current.get(receiptSignature);
        if (!receipt) {
          const target = resource === 'moderation' ? 'events' : resource;
          const engine = adminModules[resource].engine;
          const { data } = await adminRequest<{ data: AdminRecord }>(`${engine}/${target}/${row.reference}`);
          const payload = target === 'events' || target === 'spaces' ? { revision: data.revision, status: action, ...(target === 'events' ? { acknowledge_attendees: true } : {}) } : target === 'registrations' ? { updated_at: data.updated_at, status: action, notify_guest: notify, message: '' } : target === 'messaging' ? { updated_at: data.updated_at } : {};
          receipt = { path: `${engine}/${target}/${row.reference}/${target === 'messaging' ? 'retry' : target === 'tickets' ? 'check-in' : 'status'}`, body: JSON.stringify(payload), key: crypto.randomUUID() };
          if (receipts.current.size >= 40) receipts.current.clear();
          receipts.current.set(receiptSignature, receipt);
        }
        await adminRequest(receipt.path, { method: 'POST', headers: mutationHeaders(receipt.key), body: receipt.body });
        next.push({ reference: row.reference, label, message: 'Completed', success: true });
      } catch (failure) { next.push({ reference: row.reference, label, message: failure instanceof Error ? failure.message : 'The operation failed. Retry this record.', success: false }); }
      setResults([...next]);
    }
    setBusy(false); onBusyChange(false); setProgress(`${next.filter(row => row.success).length} completed; ${next.filter(row => !row.success).length} need attention.`); setConfirmed(false); onFinished();
  }
  if (!options.length || !records.length && !results.length) return null;
  return <section className="admin-bulk" aria-labelledby="bulk-actions-heading"><div className="admin-section-heading"><h2 id="bulk-actions-heading">Selected records <span className="admin-secondary-copy">{records.length}/20</span></h2><button type="button" disabled={busy} className="app-button app-button-outline app-button-sm" onClick={() => { onClear(); setResults([]); setProgress(''); }}>Clear selection</button></div><form onSubmit={submit}><fieldset disabled={busy}><RadioGroup label="Bulk action" value={action} onValueChange={value => { setAction(value); setConfirmed(false); }} options={options} disabled={busy} required />{resource === 'registrations' && <label className="campaign-confirm"><Checkbox checked={notify} onCheckedChange={setNotify} disabled={busy} />Notify affected guests by email</label>}<label className="campaign-confirm"><Checkbox required checked={confirmed} onCheckedChange={setConfirmed} disabled={busy} />Apply this action to the selected records. Each record is checked separately and changes are audited.</label><button className="app-button app-button-primary" disabled={busy || !records.length || !confirmed}>{busy ? 'Processing…' : `Apply to ${records.length} records`}</button></fieldset></form><p role="status" className="admin-secondary-copy">{progress}</p>{results.length > 0 && <ul className="admin-bulk-results">{results.map(result => <li key={result.reference}><strong>{result.label}</strong><span className={result.success ? 'admin-action-success' : 'field-error'}>{result.message}</span></li>)}</ul>}</section>;
}
