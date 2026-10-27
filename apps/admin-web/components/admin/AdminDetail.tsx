import type { ReactNode } from 'react';
import Link from 'next/link';
import { SectionJumpButton } from '@hackways/ui/section-jump';
import { Icon } from '@/components/ui/Icon';
import { type AdminModule, type AdminRecord, type AdminCell, adminModules } from '@/lib/admin-modules';
import { detailPath, isAdminModule, relatedPath } from '@/lib/admin-detail-paths';
import type { AdminDetail as DetailResponse } from '@/lib/admin-details';
import { AdminActions, BrowserBlockAction, SessionRevoke } from './AdminActions';

type Field = { key: string; label: string; date?: boolean; resource?: AdminModule; idKey?: string; url?: boolean };
type Group = { title: string; fields: Field[] };
const eventGroups: Group[] = [
  { title: 'Event details', fields: [
    { key: 'status', label: 'Publication' }, { key: 'host_name', label: 'Hosted by', resource: 'people', idKey: 'host_user_id' },
    { key: 'starts_at', label: 'Starts', date: true }, { key: 'ends_at', label: 'Ends', date: true }, { key: 'timezone', label: 'Time zone' },
    { key: 'format', label: 'Format' }, { key: 'location_name', label: 'Venue' }, { key: 'location_address', label: 'Address' },
    { key: 'online_url', label: 'Online link', url: true }, { key: 'visibility', label: 'Visibility' },
    { key: 'space_id', label: 'Space', resource: 'spaces' },
  ] },
  { title: 'Registration settings', fields: [
    { key: 'current_rsvp_count', label: 'Confirmed' }, { key: 'held_rsvp_count', label: 'Held places' }, { key: 'capacity', label: 'Capacity' },
    { key: 'registration_mode', label: 'Registration' }, { key: 'allow_guest', label: 'Guest registration' }, { key: 'waitlist_enabled', label: 'Waitlist' },
    { key: 'registration_opens_at', label: 'Opens', date: true }, { key: 'registration_closes_at', label: 'Closes', date: true },
  ] },
];
const groups: Record<AdminModule, Group[]> = {
  events: eventGroups, moderation: eventGroups,
  people: [{ title: 'Account', fields: [
    { key: 'email', label: 'Email' }, { key: 'email_verified', label: 'Email verified' }, { key: 'access', label: 'Platform access' },
    { key: 'active_sessions', label: 'Active sessions' }, { key: 'host_name', label: 'Host profile' }, { key: 'host_verified', label: 'Host verified' },
  ] }],
  spaces: [{ title: 'Space details', fields: [
    { key: 'owner_name', label: 'Owner', resource: 'people', idKey: 'owner_user_id' }, { key: 'status', label: 'Status' },
    { key: 'member_count', label: 'Members' }, { key: 'website_url', label: 'Website', url: true }, { key: 'guidelines', label: 'Guidelines' },
  ] }],
  registrations: [{ title: 'Registration', fields: [
    { key: 'event_title', label: 'Event', resource: 'events', idKey: 'event_id' }, { key: 'name', label: 'Guest', resource: 'people', idKey: 'user_id' },
    { key: 'email', label: 'Email' }, { key: 'status', label: 'Status' }, { key: 'registered_at', label: 'Registered', date: true },
    { key: 'notes', label: 'Notes' }, { key: 'offer_expires_at', label: 'Offer expires', date: true }, { key: 'approval_expires_at', label: 'Approval expires', date: true },
    { key: 'ticket_status', label: 'Ticket', resource: 'tickets', idKey: 'ticket_id' }, { key: 'transfer_count', label: 'Transfers' },
  ] }],
  tickets: [{ title: 'Ticket details', fields: [
    { key: 'event_title', label: 'Event', resource: 'events', idKey: 'event_id' }, { key: 'name', label: 'Guest', resource: 'people', idKey: 'user_id' },
    { key: 'email', label: 'Email' }, { key: 'status', label: 'Status' }, { key: 'rsvp_id', label: 'Registration', resource: 'registrations' },
    { key: 'checked_in_at', label: 'Checked in', date: true }, { key: 'checked_in_by', label: 'Checked in by', resource: 'people' },
  ] }],
  messaging: [{ title: 'Delivery details', fields: [
    { key: 'recipient_name', label: 'Recipient' }, { key: 'recipient_email', label: 'Email' }, { key: 'template_type', label: 'Template' },
    { key: 'status', label: 'Delivery' }, { key: 'delivery_state', label: 'Delivery state' }, { key: 'retry_count', label: 'Retries' },
    { key: 'scheduled_at', label: 'Scheduled', date: true }, { key: 'first_sent_at', label: 'First attempt', date: true }, { key: 'sent_at', label: 'Sent', date: true },
    { key: 'event_title', label: 'Event', resource: 'events', idKey: 'event_id' }, { key: 'registration_code', label: 'Registration', resource: 'registrations', idKey: 'rsvp_id' },
  ] }],
  storage: [{ title: 'Image details', fields: [
    { key: 'kind', label: 'Image type' }, { key: 'url', label: 'Image URL', url: true },
  ] }],
  audit: [{ title: 'Recorded action', fields: [
    { key: 'action', label: 'Action' }, { key: 'actor_name', label: 'Changed by', resource: 'people', idKey: 'actor_user_id' },
    { key: 'event_title', label: 'Event', resource: 'events', idKey: 'event_id' },
  ] }],
};

function display(value: AdminCell | undefined, field: Field): string {
  if (value === null || value === undefined || value === '') return field.key === 'capacity' ? 'No limit' : 'Not provided';
  if (typeof value === 'boolean') return value ? 'Yes' : 'No';
  if (field.date && typeof value === 'string') {
    const date = new Date(value);
    if (!Number.isNaN(date.getTime())) return new Intl.DateTimeFormat('en-IN', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Kolkata' }).format(date);
  }
  return typeof value === 'string' && ['status', 'access', 'format', 'registration_mode', 'kind', 'template_type', 'delivery_state'].includes(field.key) ? value.replaceAll('_', ' ') : String(value);
}

function valueNode(record: AdminRecord, field: Field): ReactNode {
  const value = record[field.key];
  const valueText = display(value, field);
  const referenceKey = field.key === 'checked_in_by' ? 'checked_in_by_reference' : (field.idKey || field.key).replace(/_id$/, '_reference');
  const reference = record[referenceKey];
  if (field.resource && typeof reference === 'number' && Number.isSafeInteger(reference) && reference > 0) return <a className="quiet-link" href={detailPath(field.resource, reference)}>{field.idKey ? valueText : `#${reference}`}<Icon name="arrow" size={15} /></a>;
  if (field.resource && !field.idKey) return 'Not provided';
  if (field.url && typeof value === 'string' && /^https:\/\//i.test(value)) return <a className="quiet-link admin-detail-url" href={value} rel="noreferrer" target="_blank">{field.key === 'url' ? 'Open image' : 'Open link'}<Icon name="external" size={15} /></a>;
  return valueText;
}

function titleFor(resource: AdminModule, record: AdminRecord) {
  const value = resource === 'people' ? record.display_name : resource === 'events' || resource === 'moderation' ? record.title
    : resource === 'messaging' ? record.template_type : resource === 'audit' ? record.action : resource === 'spaces' || resource === 'storage' ? record.name : record.public_code;
  return typeof value === 'string' ? resource === 'messaging' || resource === 'audit' ? value.replaceAll('_', ' ') : value : adminModules[resource].label;
}

function auditMetadata(value: AdminCell | undefined) {
  if (typeof value !== 'string') return '';
  try {
    return JSON.stringify(JSON.parse(value), (key: string, item: unknown) => {
      const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
      if (/^(?:id|.*_ids?)$/.test(key) && (typeof item === 'string' && uuid.test(item)
        || Array.isArray(item) && item.every(entry => typeof entry === 'string' && uuid.test(entry)))) return undefined;
      return item;
    }, 2);
  } catch { return ''; }
}

export function AdminDetail({ resource, detail }: { resource: AdminModule; detail: DetailResponse }) {
  const { data: record, related } = detail;
  const kind = typeof record.kind === 'string' ? record.kind : '';
  const description = typeof record.description === 'string' ? record.description : '';
  const hasPrimaryActions = ['people', 'spaces', 'storage'].includes(resource)
    || ['events', 'moderation'].includes(resource);
  const recordFields: Field[] = [{ key: 'reference', label: 'Reference' }, { key: 'created_at', label: 'Created', date: true }, { key: 'updated_at', label: 'Last changed', date: true }];
  return <>
    <a className="app-button app-button-outline app-button-sm admin-detail-back" href={`/${resource}`}><Icon name="back" size={17} />Back</a>
    <div className="admin-page-heading admin-detail-heading"><h1>{titleFor(resource, record)}</h1></div>
    {description && <p className="admin-detail-description">{description}</p>}
    {(hasPrimaryActions || related.tabs.length > 0) && <nav className="admin-detail-shortcuts" aria-label="On this record">
      {hasPrimaryActions && <SectionJumpButton targetId="record-actions" className="quiet-link">Edit &amp; actions<Icon name="arrow" size={15} /></SectionJumpButton>}
      {related.tabs.length > 0 && <SectionJumpButton targetId="record-related" className="quiet-link">Related records<Icon name="arrow" size={15} /></SectionJumpButton>}
    </nav>}
    <div className="admin-detail-layout">
      <div className="admin-detail-facts">
        {groups[resource].map(group => <section key={group.title} className="admin-detail-group"><h2>{group.title}</h2><dl>{group.fields.map(field => <div key={field.key}><dt>{field.label}</dt><dd>{valueNode(record, field)}</dd></div>)}</dl></section>)}
        {resource === 'storage' && isAdminModule(String(record.parent_resource)) && <p className="admin-detail-parent"><a className="quiet-link" href={detailPath(record.parent_resource as AdminModule, record.reference)}>View attached {record.parent_resource === 'events' ? 'event' : 'profile'}<Icon name="arrow" size={16} /></a></p>}
        {resource === 'audit' && <section className="admin-detail-group"><h2>Change details</h2><pre className="admin-audit-metadata">{auditMetadata(record.metadata)}</pre></section>}
        <section className="admin-detail-group"><h2>Record</h2><dl>{recordFields.filter(field => record[field.key] !== undefined).map(field => <div key={field.key}><dt>{field.label}</dt><dd>{valueNode(record, field)}</dd></div>)}</dl></section>
      </div>
      <div id="record-actions"><AdminActions resource={resource} record={record} /></div>
    </div>
    {related.tabs.length > 0 && <section className="admin-related" id="record-related">
      <h2>Related records</h2>
      <nav className="admin-detail-tabs" aria-label="Related records">{related.tabs.map(tab => <Link key={tab.key} href={relatedPath(resource, record.reference, tab.key, 0, kind)} scroll={false} prefetch={false} aria-current={related.section === tab.key ? 'page' : undefined}>{tab.label}</Link>)}</nav>
      {related.rows.length === 0 ? <p className="admin-related-empty">No records in this view.</p> : <div className="admin-table-scroll" role="region" aria-label={`${related.tabs.find(tab => tab.key === related.section)?.label || 'Related records'} table`} tabIndex={0}><table className="admin-table">
        <caption className="sr-only">Related records</caption><thead><tr><th scope="col">Record</th><th scope="col">Details</th><th scope="col">Status</th><th scope="col">When</th>{['sessions','devices'].includes(related.section) && <th scope="col">Action</th>}</tr></thead>
        <tbody>{related.rows.map(row => <tr key={`${row.resource}-${row.reference}`}><th scope="row">{isAdminModule(row.resource) ? <a className="admin-record-link" href={detailPath(row.resource, row.reference)}>{row.label}</a> : row.label}</th><td>{row.detail}</td><td>{row.status.replaceAll('_', ' ')}</td><td>{display(row.occurred_at, { key: 'occurred_at', label: 'When', date: true })}</td>{related.section === 'sessions' && <td>{row.status === 'active' ? <SessionRevoke person={record} sessionID={row.reference} /> : row.status === 'current' ? 'This session' : ''}</td>}{related.section === 'devices' && <td>{row.status === 'blocked' ? <Link className="quiet-link" href="/restrictions?kind=devices">Review restriction</Link> : record.access !== 'admin' ? <details><summary className="app-button app-button-outline app-button-sm">Block browser</summary><BrowserBlockAction person={record} reference={row.reference} /></details> : null}</td>}</tr>)}</tbody>
      </table></div>}
      <nav className="admin-pagination" aria-label="Related records pagination">
        {related.offset > 0 && <Link className="app-button app-button-outline app-button-sm" href={relatedPath(resource, record.reference, related.section, Math.max(0, related.offset - 20), kind)} scroll={false} prefetch={false}>Previous</Link>}
        <span>Page {Math.floor(related.offset / 20) + 1}</span>
        {related.has_next && <Link className="app-button app-button-outline app-button-sm" href={relatedPath(resource, record.reference, related.section, related.offset + 20, kind)} scroll={false} prefetch={false}>Next</Link>}
      </nav>
    </section>}
  </>;
}
