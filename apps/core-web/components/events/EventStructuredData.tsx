import type { EventDetails } from '@/types/api';
import { site } from '@/lib/site';
import { eventPath, externalUrl } from '@/lib/events';

/** Reflect only the public event projection. Meeting links and account data stay out. */
export function EventStructuredData({ event }: { event:EventDetails }) {
 if(event.visibility!=='public'||event.status!=='published')return null;
 const website=externalUrl(event.organizer_website),url=`${site.app}${eventPath(event)}`;
 const data={
  '@context':'https://schema.org','@type':'Event',name:event.title,description:event.short_description||event.description,url,
  startDate:event.starts_at,endDate:event.ends_at,eventStatus:'https://schema.org/EventScheduled',
  eventAttendanceMode:`https://schema.org/${event.format==='online'?'OnlineEventAttendanceMode':event.format==='hybrid'?'MixedEventAttendanceMode':'OfflineEventAttendanceMode'}`,
  ...(event.banner_url?{image:[event.banner_url]}:{}),
  ...(event.format!=='online'&&(event.location_name||event.location_address)?{location:{'@type':'Place',...(event.location_name?{name:event.location_name}:{}),...(event.location_address?{address:event.location_address}:{})}}:{}),
  organizer:event.space_slug?{'@type':'Organization',name:event.space_name,url:`${site.app}/spaces/${event.space_slug}`}:{'@type':'Organization',name:event.organizer_name,...(website?{url:website}:{})},
 };
 return <script type="application/ld+json" dangerouslySetInnerHTML={{__html:JSON.stringify(data).replace(/</g,'\\u003c')}} />;
}
