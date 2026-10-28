import type { ReactNode } from 'react';
import { brandIcons } from '@hackways/ui/brand-metadata';
import Link from 'next/link';
import { readEvent } from '@/lib/event-details';

import { AmbientBackdrop } from '@/components/events/AmbientBackdrop';
import { Icon } from '@/components/ui/Icon';
import { AppNavigation } from '@/components/ui/AppNavigation';
import { appearanceOf, appearanceStyle } from '@/lib/appearance';
import '@/components/hosting/appearance.css';
import './event-details.css';

export const metadata = { icons: brandIcons('', 'event') };

export default async function EventLayout({ children, params }: { children: ReactNode; params: { slug: string } }) {
  const event = await readEvent(params.slug);

  const appearance = appearanceOf(event);
  return <>
    <div className={`event-page-backdrop event-theme theme-${appearance.theme} event-display-${appearance.theme_display}`} style={appearanceStyle(appearance)} aria-hidden="true"><AmbientBackdrop active={appearance.theme === 'ambient'} source={event.banner_url} /></div>
    <div className={`event-page event-scene event-theme theme-${appearance.theme} event-display-${appearance.theme_display}`} style={appearanceStyle(appearance)}>
    <header className="event-mobile-header">
      <Link className="event-mobile-brand" href="/discover" aria-label="Hackways home">h.</Link>
      <AppNavigation compact />
      <div className="event-mobile-account-actions"><Link className="event-mobile-create" href="/create" aria-label="Create event" title="Create event"><Icon name="add" size={20}/></Link><Link className="event-mobile-account" href="/account" aria-label="Account" title="Account"><Icon name="account" size={20} /></Link></div>
    </header>
    <div className="event-scene-inner"><Link className="event-back" href="/discover" aria-label="Back to Discover events"><Icon name="arrow" size={16} className="reverse-icon" /><span className="event-back-desktop">Discover events</span><span className="event-back-mobile">Back</span></Link>{children}</div>
  </div></>;
}
