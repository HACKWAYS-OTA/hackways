import Link from 'next/link';
import type { EventDetails } from '@/types/api';
import { eventPath, registrationState } from '@/lib/events';
import { Icon } from '@/components/ui/Icon';
import { RefreshEvent } from './RefreshEvent';
import type { EventManagementAccess } from '@/lib/event-details';
import { site } from '@/lib/site';

export function EventInvitation({ event, managementAccess = 'denied', invited = false, registrationCode = null }: { event: EventDetails; managementAccess?: EventManagementAccess; invited?:boolean; registrationCode?:string|null }) {
  const registration = registrationState(event);
  return <div className="event-invitation">
    {registrationCode ? <Link className="app-button app-button-primary" href={`/my-events/${registrationCode}`}>View registration<Icon name="arrow" size={18} /></Link> : registration.disabled ? <div className="event-invitation-closed"><p>{registration.label}</p></div> : <Link className="app-button app-button-primary" href={`${eventPath(event)}/register`} scroll={false}>{invited ? 'Accept invitation' : registration.label === 'Join waiting list' ? registration.label : 'Register'}<Icon name="arrow" size={18} /></Link>}
    {managementAccess === 'allowed' && <a className="app-button app-button-outline" href={`${site.organizer}${eventPath(event)}`}>Manage event<Icon name="external" size={17} /></a>}
    {managementAccess === 'unavailable' && <div className="event-management-error" role="status"><span>Management access could not be checked.</span><RefreshEvent label="Retry access check" /></div>}
  </div>;
}
