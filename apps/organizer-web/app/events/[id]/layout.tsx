import Link from 'next/link';
import type { ReactNode } from 'react';
import { managedEventContext } from '@/lib/management';
import { dateLabel } from '@/lib/journeys';
import { site } from '@/lib/site';
import { eventStatus } from '@/lib/dashboard';
import { ManagementTabs } from '@/components/management/ManagementTabs';

import { Icon } from '@/components/ui/Icon';
import { ManagementChrome } from '@/components/management/ManagementChrome';
export default async function EventLayout({ params, children }: { params: { id: string }; children: ReactNode }) {
  const { data: event, access } = await managedEventContext(params.id);
  const editable = !event.management_closed && access.permissions.includes('content');
  return <div className="host-workspace"><ManagementChrome><div className="host-back-row"><Link className="app-button app-button-outline app-button-sm" href="/"><Icon name="arrow" size={15} className="reverse-icon" />All events</Link></div><header className="host-event-heading"><div>{event.status !== 'published' && <p className="host-event-state">{eventStatus(event)}</p>}<h1>{event.title}</h1><p><Icon name="calendar" size={16} />{dateLabel(event.starts_at, event.timezone)}<span>{event.timezone}</span></p></div><div className="host-event-heading-actions">{editable && <Link className="app-button app-button-primary" href={`/events/${event.slug}/edit`}>Edit event</Link>}{event.status === 'published' && <a className="app-button app-button-outline" href={`${site.app}/events/${event.slug}`} target="_blank" rel="noopener noreferrer">View event<Icon name="external" size={16} /></a>}</div></header><ManagementTabs slug={event.slug} permissions={access.permissions} /></ManagementChrome><div className="host-workspace-content">{children}</div></div>;
}
