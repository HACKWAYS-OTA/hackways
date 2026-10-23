import { managedEvent } from '@/lib/management';
import { EventAppearanceEditor } from '@/components/management/EventAppearanceEditor';
import { EventPageTabs } from '@/components/management/EventPageTabs';
export const metadata = { title: 'Event cover and theme' };
export default async function Appearance({ params }: { params: { id: string } }) {
  const event = await managedEvent(params.id);
  return <><EventPageTabs slug={event.slug} appearance />{event.management_closed ? <p className="journey-notice">An administrator has closed editing for this event.</p> : <EventAppearanceEditor key={event.revision} event={event} />}</>;
}
