import Link from 'next/link';
import { notFound } from 'next/navigation';
import { requireUser, privateApi } from '@/lib/account';
import { ApiClientError } from '@/lib/api';
import type { Event, Registration } from '@/types/api';
import { dateLabel, registrationLabels } from '@/lib/journeys';
import { RegistrationControls } from '@/components/account/RegistrationControls';
export const metadata = { title: 'Your registration', robots: { index: false, follow: false } };
export default async function RegistrationPage({ params }: { params: { id: string } }) {
  if (!/^R-[0-9a-f]{16}$/.test(params.id)) notFound();
  await requireUser(`/my-events/${params.id}`);
  let registration: Registration;
  try { registration = (await privateApi<{ data: Registration }>(`/rsvps/by-code/${params.id}`)).data; }
  catch (error) { if (error instanceof ApiClientError && error.code === 'NOT_FOUND') notFound(); throw error; }
  // An unpublished event can become unavailable; the persisted registration still remains accessible.
  let event: Event | null = null;
  try { event = (await privateApi<{ data: Event }>(`/events/${registration.event_id}`)).data; }
  catch (error) { if (!(error instanceof ApiClientError && error.code === 'NOT_FOUND')) throw error; }
  let pass: { qr_image: string; short_code: string } | null = null;
  let passError = '';
  if (registration.status === 'confirmed' && event?.status === 'published') {
    try { pass = await privateApi<{ qr_image: string; short_code: string }>(`/rsvps/${registration.id}/ticket`); }
    catch (error) { if (error instanceof ApiClientError) passError = error.message; else throw error; }
  }
  return <div className="account-page"><Link className="quiet-link" href="/my-events">Back to my events</Link><div className="registration-heading"><span className={`journey-status status-${registration.status}`}>{registrationLabels[registration.status]}</span><h1>{event?.title || 'Your registration'}</h1>{event ? <p>{dateLabel(event.starts_at, event.timezone)} · {event.timezone}</p> : <p>The event page is currently unavailable.</p>}</div><RegistrationControls registration={registration} startsAt={event?.starts_at} initialTicket={pass} initialTicketError={passError} />{event && <Link className="quiet-link" href={`/events/${encodeURIComponent(event.slug)}`}>See event details</Link>}</div>;
}
