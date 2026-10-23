import { managedEvent } from '@/lib/management';
import { EventEditWorkspace } from '@/components/management/EventEditWorkspace';
import '@/components/hosting/appearance.css';
import '@/components/management/event-edit.css';
export const metadata = { title: 'Edit event details' };
export default async function Edit({ params }: { params: { id: string } }) {
  const event = await managedEvent(params.id);
  if (event.management_closed) return <section className="desk-empty"><h2>These details are now closed.</h2><p>An administrator has closed editing for this event. You can still view the guest list.</p></section>;
  return <EventEditWorkspace event={event} timeZones={Intl.supportedValuesOf('timeZone')} />;
}
