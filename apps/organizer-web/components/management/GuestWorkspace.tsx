import Link from 'next/link';
import { privateApi, pageOffset } from '@/lib/account';
import { queryPath, type GuestSummary } from '@/lib/dashboard';
import type { Guest } from '@/lib/guests';
import type { ManagedEvent } from '@/lib/management';
import { GuestList } from './GuestList';
import { CheckInForm } from './CheckInForm';
import { Icon } from '@/components/ui/Icon';
export async function GuestWorkspace({ event, searchParams, checkIn = false }: { event: ManagedEvent; searchParams: { offset?: string; q?: string; status?: string }; checkIn?: boolean }) {
  const choices = checkIn ? [['not_arrived', 'Not arrived'], ['checked_in', 'Checked in'], ['confirmed', 'Everyone going']] : [['all', 'All guests'], ['confirmed', 'Going'], ['pending_approval', 'Awaiting approval'], ['waitlisted', 'Waiting list'], ['offered', 'Offered a place'], ['cancelled', 'Cancelled'], ['rejected', 'Declined'], ['expired', 'Expired'], ['transferred', 'Transferred']];
  const status = choices.find(([value]) => value === searchParams.status)?.[0] || choices[0][0];
  const q = typeof searchParams.q === 'string' ? searchParams.q.trim().slice(0, 200) : '';
  const offset = pageOffset(searchParams.offset), path = `/events/${event.slug}/${checkIn ? 'check-in' : 'guests'}`;
  const [roster, summary] = await Promise.allSettled([privateApi<{ data: Guest[] }>(queryPath(`/events/${event.id}/registrations`, { q, status, offset, limit: 26 })), privateApi<{ data: GuestSummary }>(`/events/${event.id}/summary`)]);
  const open = event.status === 'published' && Date.now() >= Date.parse(event.starts_at) - 86400000 && Date.now() < Date.parse(event.ends_at);
  const totals = summary.status === 'fulfilled' ? summary.value.data : null;
  const summaryLinks = totals ? checkIn
    ? [['checked_in', 'Checked in', totals.checked_in], ['not_arrived', 'Yet to arrive', Math.max(0, totals.confirmed - totals.checked_in)], ['confirmed', 'Everyone going', totals.confirmed]] as const
    : [['confirmed', 'Going', totals.confirmed], ['pending_approval', 'Awaiting approval', totals.pending], ['waitlisted', 'Waiting list', totals.waitlisted], ['offered', 'Offered a place', totals.offered]] as const
    : [];
  return <>{checkIn && <CheckInForm event={event} open={open} />}
    {totals && <nav className="host-guest-summary" aria-label="Guest totals and filters">{summaryLinks.map(([value, label, count]) => <Link key={value} href={queryPath(path, { status: value, q })} aria-current={status === value ? 'page' : undefined}><span>{label}</span><strong>{count}</strong><Icon name="arrow" size={16} /></Link>)}</nav>}
    {!checkIn && <div className="host-roster-tools"><Link className="app-button app-button-primary" href={`/events/${event.slug}/invitations`}><Icon name="add" size={16} />Invite attendees</Link><Link className="app-button app-button-outline" href={`/events/${event.slug}/check-in`}><Icon name="check" size={16} />Check in guests</Link><Link className="quiet-link" href={`/events/${event.slug}/registration/questions`}>Registration questions<Icon name="arrow" size={15} /></Link></div>}
    <div className="host-section-heading"><h2>{checkIn ? 'Find your guest' : 'Guest list'}</h2><form key={`${status}-${q}`} className="host-search" role="search"><input type="hidden" name="status" value={status} /><Icon name="search" size={17} /><input type="search" name="q" defaultValue={q} maxLength={200} aria-label="Search guests by name, email or registration reference" placeholder="Name, email or reference" /><button type="submit">Search</button></form></div>
    <nav className="host-guest-filters" aria-label="Guest status">{choices.map(([value, label]) => <Link key={value} href={queryPath(path, { status: value, q })} aria-current={value === status ? 'page' : undefined}>{label}</Link>)}</nav>
    {roster.status === 'fulfilled' ? <GuestList key={`${status}-${q}-${offset}`} event={event} guests={roster.value.data.slice(0, 25)} checkIn={checkIn} checkinOpen={open} /> : <section className="host-empty"><h2>Your guest list couldn’t be loaded.</h2><p role="alert">{roster.reason instanceof Error ? roster.reason.message : 'Please try again.'}</p><Link className="app-button app-button-outline" href={queryPath(path, { status, q, offset })}>Try again</Link></section>}
    {summary.status === 'rejected' && <p className="field-help" role="status">Guest totals are temporarily unavailable. Refresh to try again.</p>}
    {roster.status === 'fulfilled' && (offset > 0 || roster.value.data.length > 25) && <nav className="journey-pagination" aria-label="More guests">{offset > 0 && <Link className="app-button app-button-outline" href={queryPath(path, { status, q, offset: Math.max(0, offset - 25) })}>Previous</Link>}{roster.value.data.length > 25 && <Link className="app-button app-button-outline" href={queryPath(path, { status, q, offset: offset + 25 })}>Next guests</Link>}</nav>}
  </>;
}
