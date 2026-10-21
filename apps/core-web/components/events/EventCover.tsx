"use client";
import { useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import type { Event } from '@/types/api';
import { EventPoster } from './EventPoster';

export function EventCover({ event, priority = false, className = 'catalog-event-cover', href, sizes = '(max-width: 640px) 100vw, 440px' }: { event: Event; priority?: boolean; className?: string; href?: string; sizes?: string }) {
  const [failedSource, setFailedSource] = useState<string | null>(null);
  let source: string | null = null;
  try { const url = new URL(event.banner_url || ''); if (url.protocol === 'https:' && !url.username && !url.password) source = url.href; } catch { /* Events without a cover use their branded poster. */ }
  const visual = source && source !== failedSource
    ? <Image className="catalog-event-image" src={source} alt="" fill sizes={sizes} priority={priority} unoptimized onError={() => setFailedSource(source)} />
    : <EventPoster event={event} />;
  return href ? <Link href={href} prefetch={false} className={className} aria-label={`View ${event.title}`}>{visual}</Link> : <div className={className} aria-hidden="true">{visual}</div>;
}
