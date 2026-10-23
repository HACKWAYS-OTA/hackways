import type { Metadata } from 'next';
import Link from 'next/link';
import { readEvent, readRegistrationForm,readInvitationState } from '@/lib/event-details';
import { eventPath } from '@/lib/events';
import { RegistrationForm } from '@/components/events/RegistrationForm';
import { EventCover } from '@/components/events/EventCover';
import { currentUser } from '@/lib/account';
export async function generateMetadata({ params }: { params: { slug: string } }): Promise<Metadata> {
  const event = await readEvent(params.slug);
  return { title: `Register for ${event.title}`, robots: { index: false, follow: true } };
}
export default async function RegisterPage({ params,searchParams }: { params: { slug: string };searchParams:{invitation?:string} }) {
  const [event, user, form,invitation] = await Promise.all([readEvent(params.slug), currentUser(), readRegistrationForm(params.slug),readInvitationState(params.slug)]);
  return <div className="event-checkout registration-checkout"><header className="registration-event-summary"><EventCover event={event} className="registration-event-banner" sizes="(max-width: 640px) calc(100vw - 40px), 240px" /><div><h1>{event.title}</h1>{(event.short_description || event.description) && <p>{event.short_description || event.description}</p>}</div></header><section className="event-checkout-form" id="registration-form-start" tabIndex={-1}><nav className="event-breadcrumb" aria-label="Event navigation"><Link href={eventPath(event)}>Event details</Link><span aria-hidden="true">/</span><span aria-current="page">Registration</span></nav><RegistrationForm event={event} user={user} initialForm={form} invited={invitation.invited} invitationHint={searchParams.invitation==='1'} existingCode={invitation.registration_code} /></section></div>;
}
