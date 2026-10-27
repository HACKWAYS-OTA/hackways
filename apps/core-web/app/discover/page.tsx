import { managedMetadata } from '@hackways/ui/seo';
import { staticSocialMetadata } from '@hackways/ui/social-image';
import { site } from '@/lib/site';
import type { Metadata } from 'next';
import { EventBrowser } from '@/components/events/EventBrowser';
import { listEvents, PAGE_SIZE, type InitialEvents } from '@/lib/events';
import { ApiClientError } from '@/lib/api';
import { listPromotions } from '@/lib/promotions';
import { runtimeFeatures } from '@/lib/runtime';
const searchTitle = 'Hackathons, Meetups & Tech Events';
const searchDescription = 'Find published hackathons, developer meetups, workshops and conferences. Explore AI, software, data, cloud, cybersecurity, startups and community events.';
const pageDefaults: Metadata = { ...staticSocialMetadata(site.app, 'discover', searchTitle, searchDescription), title: searchTitle, description: searchDescription, alternates: { canonical: '/discover' } };
export function generateMetadata() { return managedMetadata("/discover", pageDefaults); }
export default async function DiscoverPage() {
  let initial: InitialEvents;
  const [events, promotions, flags] = await Promise.allSettled([listEvents(), listPromotions(), runtimeFeatures()]);
  if (events.status === 'fulfilled') initial = { events: events.value.data, error: null, hasMore: events.value.data.length === PAGE_SIZE };
  else initial = { events: [], error: events.reason instanceof ApiClientError ? events.reason.message : 'Events could not be loaded. Please try again.', hasMore: false };
  return <EventBrowser initial={initial} promotions={promotions.status === 'fulfilled' && flags.status === 'fulfilled' && flags.value['discover.promotions'] ? promotions.value : null} />;
}
