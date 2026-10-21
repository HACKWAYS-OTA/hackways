import type { Metadata } from 'next';
import Link from 'next/link';
import { readEvent, readRegistrationForm } from '@/lib/event-details';
import { eventPath } from '@/lib/events';
import { RegistrationForm } from '@/components/events/RegistrationForm';
import { EventCover } from '@/components/events/EventCover';
import { EventSidebar } from '@/components/events/EventSidebar';
import { currentUser } from '@/lib/account';
export async function generateMetadata({ params }: { params: { slug: string } }): Promise<Metadata> {
  const event = await readEvent(params.slug);
  return { title: `Register for ${event.title}`, robots: { index: false, follow: true } };
}
export default async function RegisterPage({ params }: { params: { slug: string } }) {
  const [event, user, form] = await Promise.all([readEvent(params.slug), currentUser(), readRegistrationForm(params.slug)]);
  return <div className="event-checkout"><section className="event-checkout-form"><nav className="event-breadcrumb" aria-label="Event navigation"><Link href={eventPath(event)}>Event details</Link><span aria-hidden="true">/</span><span aria-current="page">Registration</span></nav><RegistrationForm event={event} user={user} initialForm={form} /></section><aside className="event-checkout-aside" aria-label="Event summary"><EventCover event={event} className="event-hero-image" sizes="(max-width: 860px) calc(100vw - 40px), 39vw" /><EventSidebar event={event} /></aside></div>;
}
