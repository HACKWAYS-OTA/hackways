'use client';

import { useRef, useState, type ReactNode, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { RadioGroup } from '@hackways/ui/radio-group';
import { Checkbox } from '@hackways/ui/checkbox';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { type AdminModule, type AdminRecord } from '@/lib/admin-modules';

type Payload = Record<string, string | number | boolean | null>;
const text = (record: AdminRecord, key: string) => typeof record[key] === 'string' ? record[key] as string : '';
const number = (record: AdminRecord, key: string) => typeof record[key] === 'number' ? record[key] as number : 0;

function ActionForm({ title, description, path, method = 'POST', payload, fields, children, submitLabel, success = 'Changes saved.', destination }: {
  title: string; description?: string; path: string; method?: 'POST' | 'PUT'; payload: Payload;
  fields?: (data: FormData) => Payload; children?: ReactNode; submitLabel: string; success?: string; destination?: string;
}) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);
  const receipt = useRef<{ signature: string; key: string } | null>(null);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy) return;
    const values = new FormData(event.currentTarget);
    const reason = String(values.get('admin_reason') || '').trim();
    const body = JSON.stringify({ ...payload, ...fields?.(values) });
    const signature = `${path}\n${reason}\n${body}`;
    if (receipt.current?.signature !== signature) receipt.current = { signature, key: crypto.randomUUID() };
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), 15000);
    setBusy(true); setError(''); setSaved(false);
    try {
      const base = (process.env.NEXT_PUBLIC_API_URL || '/api/v1').replace(/\/$/, '');
      const response = await fetch(`${base}/admin/${path}`, {
        method, credentials: 'include', cache: 'no-store', signal: controller.signal,
        headers: { 'Content-Type': 'application/json', 'Idempotency-Key': receipt.current.key, ...(reason ? { 'X-Admin-Reason': encodeURIComponent(reason) } : {}) },
        body,
      });
      if (response.status === 401) { window.location.assign(`/login?next=${encodeURIComponent(window.location.pathname + window.location.search)}`); return; }
      if (!response.ok) {
        const failure: unknown = await response.json().catch(() => null);
        const envelope = failure && typeof failure === 'object' && 'error' in failure ? failure.error : null;
        if (envelope && typeof envelope === 'object' && 'code' in envelope && envelope.code === 'ADMIN_VERIFICATION_REQUIRED') { window.location.assign(`/verify?next=${encodeURIComponent(window.location.pathname + window.location.search)}`); return; }
        if (response.status === 403 && envelope && typeof envelope === 'object' && 'code' in envelope && envelope.code === 'ADMIN_ACCESS_REQUIRED') {
          window.location.replace('/access-denied');
          return;
        }
        throw new Error(envelope && typeof envelope === 'object' && 'message' in envelope && typeof envelope.message === 'string'
          ? envelope.message : 'This change could not be saved. Try again.');
      }
      setSaved(true);
      if (destination) { window.location.assign(destination); return; }
      router.refresh();
    } catch (failure) {
      setError(failure instanceof Error && failure.name !== 'AbortError' ? failure.message : 'The request timed out. Retry to check the same change safely.');
    } finally { window.clearTimeout(timeout); setBusy(false); }
  }
  return <section className="admin-action-section">
    <h2>{title}</h2>
    {description && <p className="admin-action-description">{description}</p>}
    <form onSubmit={submit} className="admin-action-form">
      <fieldset disabled={busy}>
        {children}
        {error && <p className="field-error" role="alert">{error}</p>}
        {saved && <p className="admin-action-success" role="status">{success}</p>}
        <Button type="submit" isLoading={busy}>{busy ? 'Saving…' : submitLabel}</Button>
      </fieldset>
    </form>
  </section>;
}

function Choice({ name, label, value: initialValue, options }: { name: string; label: string; value: string; options: {value: string; label: string}[] }) {
  const [value, setValue] = useState(initialValue);
  return <div className="field admin-action-choice"><p>{label}</p><RadioGroup name={name} label={label} required value={value} onValueChange={setValue} options={options} layout="stack" /></div>;
}

function TextArea({ name, label, value = '', maxLength = 10000 }: { name: string; label: string; value?: string; maxLength?: number }) {
  return <div className="field"><label htmlFor={`edit-${name}`}>{label}</label><textarea id={`edit-${name}`} className="app-input" name={name} defaultValue={value} rows={5} maxLength={maxLength} /></div>;
}

function Confirmation({ name, children }: { name: string; children: ReactNode }) {
  return <label className="admin-action-confirm"><Checkbox name={name} value="yes" required />{children}</label>;
}

export function SessionRevoke({ person, sessionID }: { person: AdminRecord; sessionID: string }) {
  return <details className="admin-session-action"><summary className="app-button app-button-outline app-button-sm">Revoke session</summary><ActionForm title="Revoke this session" path={`global/people/${person.reference}/sessions/${sessionID}/revoke`} payload={{ updated_at: text(person, 'updated_at') }} submitLabel="Revoke session" success="Session revoked." /></details>;
}

export function EventBinAction({ record, restore = false }: { record: AdminRecord; restore?: boolean }) {
  return <ActionForm title={restore ? 'Restore event' : 'Move to Bin'}
    description={restore ? 'Published events return as unpublished. Cancelled registrations stay cancelled.' : 'Cancel active events first. Bookings, tickets and history are retained for recovery.'}
    path={restore ? `core/bin/events/${record.reference}/restore` : `core/events/${record.reference}/bin`}
    payload={{ revision: number(record, 'revision') }} submitLabel={restore ? 'Restore event' : 'Move to Bin'}
    destination={restore ? `/events/details/${record.reference}` : '/bin'}>
    <Input label="Reason" name="admin_reason" required minLength={3} maxLength={1000} />
    <Confirmation name="confirm_bin">{restore ? 'Restore this event without changing its bookings.' : 'Move this event out of the active catalogue.'}</Confirmation>
  </ActionForm>;
}

export function AdminActions({ resource, record }: { resource: AdminModule; record: AdminRecord }) {
  const version = text(record, 'updated_at');
  const revision = number(record, 'revision');
  const status = text(record, 'status');
  const id = record.reference;
  if (resource === 'audit') return null;
  if (resource === 'people') return <div className="admin-actions" key={version}>
    <ActionForm title="Edit profile" path={`global/people/${id}/edit`} payload={{ updated_at: version }} fields={data => ({ display_name: String(data.get('display_name') || '') })} submitLabel="Save profile">
      <Input label="Display name" name="display_name" defaultValue={text(record, 'display_name')} required maxLength={100} />
    </ActionForm>
    <ActionForm title="Admin access" description="Admin access allows this account to view and manage platform records." path={`global/people/${id}/access`} payload={{ updated_at: version }} fields={data => ({ access: String(data.get('access') || '') })} submitLabel="Update access">
      <Choice name="access" label="Access" value={text(record, 'access')} options={[{ value: 'member', label: 'Member' }, { value: 'admin', label: 'Admin' }]} />
      <Confirmation name="confirm_access">I have checked this account and the access change.</Confirmation>
    </ActionForm>
  </div>;
  if (resource === 'events' || resource === 'moderation') {
    const transitions = [
      ...(record.editable === true && status !== 'published' ? [{ value: 'published', label: 'Publish event' }] : []),
      ...(status === 'published' ? [{ value: 'unpublished', label: 'Unpublish event' }] : []),
      ...(status !== 'cancelled' ? [{ value: 'cancelled', label: 'Cancel event' }] : []),
    ];
    return <div className="admin-actions" key={String(revision)}>
      {record.editable === true && <ActionForm title="Edit event" method="PUT" path={`core/events/${id}`} payload={{ revision }} fields={data => ({ title: String(data.get('title') || ''), description: String(data.get('description') || ''), capacity: String(data.get('capacity') || '').trim() ? Number(data.get('capacity')) : null, visibility: String(data.get('visibility') || '') })} submitLabel="Save event">
        <Input label="Event name" name="title" defaultValue={text(record, 'title')} required maxLength={160} />
        <TextArea name="description" label="Description" value={text(record, 'description')} />
        <Input label="Capacity" name="capacity" type="number" min={Math.max(1, number(record, 'current_rsvp_count') + number(record, 'held_rsvp_count'))} max={2147483647} defaultValue={record.capacity === null ? '' : String(record.capacity)} helperText="Leave empty for no limit." />
        <Choice name="visibility" label="Visibility" value={text(record, 'visibility')} options={[{ value: 'public', label: 'Public' }, { value: 'private', label: 'Private' }]} />
      </ActionForm>}
      {transitions.length > 0 && <ActionForm title="Publication" path={`core/events/${id}/status`} payload={{ revision }} fields={data => ({ status: String(data.get('event_status') || ''), acknowledge_attendees: data.get('acknowledge_attendees') === 'yes' })} submitLabel="Apply event action">
        <Choice name="event_status" label="Event action" value={transitions[0].value} options={transitions} />
        <Confirmation name="acknowledge_attendees">I understand this changes event availability and may affect registered guests.</Confirmation>
      </ActionForm>}
      <EventBinAction record={record} />
    </div>;
  }
  if (resource === 'spaces') return <div className="admin-actions" key={String(revision)}>
    <ActionForm title="Edit space" method="PUT" path={`core/spaces/${id}`} payload={{ revision }} fields={data => ({ name: String(data.get('name') || ''), description: String(data.get('description') || ''), guidelines: String(data.get('guidelines') || ''), website_url: String(data.get('website_url') || '') })} submitLabel="Save space">
      <Input label="Space name" name="name" required maxLength={120} defaultValue={text(record, 'name')} />
      <TextArea label="Description" name="description" value={text(record, 'description')} />
      <TextArea label="Guidelines" name="guidelines" value={text(record, 'guidelines')} />
      <Input label="Website" name="website_url" type="url" defaultValue={text(record, 'website_url')} maxLength={2048} />
    </ActionForm>
    <ActionForm title={status === 'archived' ? 'Restore space' : 'Archive space'} path={`core/spaces/${id}/status`} payload={{ revision, status: status === 'archived' ? 'active' : 'archived' }} submitLabel={status === 'archived' ? 'Restore space' : 'Archive space'}>
      <Confirmation name="confirm_space">Apply this change to the space.</Confirmation>
    </ActionForm>
  </div>;
  if (resource === 'registrations') {
    const transitions = [
      { value: 'confirmed', label: 'Going' },
      { value: 'pending_approval', label: 'Awaiting approval' },
      { value: 'waitlisted', label: 'Waiting list' },
      { value: 'rejected', label: 'Declined' },
      { value: 'cancelled', label: 'Cancelled' },
    ].filter(option => option.value !== status && (option.value !== 'confirmed' || record.registration_open === true));
    if (!record.editable || record.ticket_status === 'scanned' || transitions.length === 0) return null;
    return <div className="admin-actions" key={version}><ActionForm title="Registration action" path={`core/registrations/${id}/status`} payload={{ updated_at: version }} fields={data => ({ status: String(data.get('registration_status') || ''), notify_guest: data.get('notify_guest') === 'yes', message: String(data.get('message') || '') })} submitLabel="Update registration">
      <Choice name="registration_status" label="Action" value={transitions[0].value} options={transitions} />
      <label className="admin-action-confirm"><Checkbox name="notify_guest" value="yes" defaultChecked />Notify guest by email</label>
      <TextArea name="message" label="Message to guest" maxLength={2000} />
    </ActionForm></div>;
  }
  if (resource === 'tickets') return record.can_check_in === true ? <div className="admin-actions"><ActionForm title="Check in guest" path={`core/tickets/${id}/check-in`} payload={{}} submitLabel="Check in guest" success="Guest checked in.">
    <Confirmation name="confirm_checkin">I have verified this guest and their ticket.</Confirmation>
  </ActionForm></div> : null;
  if (resource === 'messaging') return record.can_retry === true ? <div className="admin-actions" key={version}><ActionForm title="Retry delivery" description="Schedule another attempt for this failed email." path={`global/messaging/${id}/retry`} payload={{ updated_at: version }} submitLabel="Schedule retry" success="Retry scheduled." /></div> : null;
  if (resource === 'storage') return <div className="admin-actions" key={version}><ActionForm title="Remove attached image" description="Remove this image from its event or profile. The stored file is retained." path={`global/storage/${id}/detach`} payload={{ updated_at: version, kind: text(record, 'kind') }} submitLabel="Remove image" success="Image removed." destination="/storage">
    <Confirmation name="confirm_detach">Remove this image from the record.</Confirmation>
  </ActionForm></div>;
  return null;
}
