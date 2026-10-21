import { managedEvent } from '@/lib/management';
import { EventAppearanceEditor } from '@/components/management/EventAppearanceEditor';
import { EventPageTabs } from '@/components/management/EventPageTabs';
export const metadata = { title: 'Event cover and theme' };
export default async function Appearance({ params }: { params: { id: string } }) {
  const event = await managedEvent(params.id);
  return <><EventPageTabs slug={event.slug} appearance />{event.status === 'cancelled' || Date.parse(event.starts_at) <= Date.now() ? <p className="journey-notice">Appearance changes are closed for started or cancelled events.</p> : <EventAppearanceEditor key={event.revision} event={event} />}</>;
}
