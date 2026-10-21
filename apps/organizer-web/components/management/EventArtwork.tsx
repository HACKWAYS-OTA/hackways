'use client';
import { useState } from 'react';
import Image from 'next/image';
import type { Event } from '@/types/api';
import { safeLink } from '@/lib/dashboard';
import { dateLabel } from '@/lib/journeys';
export function EventArtwork({ event, compact = false }: { event: Event; compact?: boolean }) {
  const [failed, setFailed] = useState('');
  const source = safeLink(event.banner_url);
  return <div className={`host-artwork${compact ? ' host-artwork-compact' : ''}`} aria-hidden="true">{source && source !== failed ? <Image src={source} alt="" fill unoptimized sizes={compact ? '100px' : '(max-width: 700px) 90vw, 360px'} onError={() => setFailed(source)} /> : <div className="host-poster"><span>{dateLabel(event.starts_at, event.timezone, { month: 'short', day: 'numeric' })}</span><strong>{event.title}</strong><span>hackways.</span></div>}</div>;
}
