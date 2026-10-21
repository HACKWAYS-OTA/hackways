import Link from 'next/link';
import type { EventDetails } from '@/types/api';
import { eventPath, registrationState } from '@/lib/events';
import { Icon } from '@/components/ui/Icon';
import { RefreshEvent } from './RefreshEvent';
import type { EventManagementAccess } from '@/lib/event-details';
import { site } from '@/lib/site';

export function EventInvitation({ event, managementAccess = 'denied' }: { event: EventDetails; managementAccess?: EventManagementAccess }) {
  const registration = registrationState(event);
  return <div className="event-invitation">
    {registration.disabled ? <div className="event-invitation-closed"><p>{registration.label}</p><RefreshEvent /></div> : <Link className="app-button app-button-primary" href={`${eventPath(event)}/register`}>{registration.label === 'Join waiting list' ? registration.label : event.registration_mode === 'approval' ? 'Apply to join' : 'Register'}<Icon name="arrow" size={18} /></Link>}
    {managementAccess === 'allowed' && <a className="app-button app-button-outline" href={`${site.organizer}${eventPath(event)}`}>Manage event<Icon name="external" size={17} /></a>}
    {managementAccess === 'unavailable' && <div className="event-management-error" role="status"><span>Management access could not be checked.</span><RefreshEvent label="Retry access check" /></div>}
  </div>;
}
