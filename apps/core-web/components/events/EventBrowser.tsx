"use client";

import { useCallback, useEffect, useRef, useState } from 'react';
import { listEvents, PAGE_SIZE, type InitialEvents } from '@/lib/events';
import Link from 'next/link';
import { site } from '@/lib/site';
import { EventCard } from './EventCard';
import { Button } from '@/components/ui/Button';
import { Icon } from '@/components/ui/Icon';
import { EventCardSkeleton } from '@/components/ui/Skeleton';
import './discovery-reading.css';
import { DiscoverPromotions } from './DiscoverPromotions';
import type { PromotionFeed } from '@/lib/promotions';

export function EventBrowser({ initial, promotions }: { initial: InitialEvents; promotions: PromotionFeed | null }) {
  const [events, setEvents] = useState(initial.events);
  const [error, setError] = useState(initial.error);
  const [pageError, setPageError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(initial.hasMore);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [query, setQuery] = useState('');
  const [appliedQuery, setAppliedQuery] = useState('');
  const searchTimer = useRef<ReturnType<typeof setTimeout>>();
  const controller = useRef<AbortController>();
  const offset = useRef(initial.events.length);
  const requestId = useRef(0);
  const completedQuery = useRef('');

  const request = useCallback(async (search: string, append = false) => {
    controller.current?.abort();
    const abort = new AbortController();
    controller.current = abort;
    const id = ++requestId.current;
    const start = append ? offset.current : 0;
    if (append) setLoadingMore(true); else setLoading(true);
    setPageError(null);
    try {
      const page = await listEvents(search, start, abort.signal);
      if (abort.signal.aborted || id !== requestId.current) return;
      if (!append) setError(null);
      setEvents((previous) => append ? Array.from(new Map([...previous, ...page.data].map((event) => [event.id, event])).values()) : page.data);
      offset.current = start + page.data.length;
      setHasMore(page.data.length === PAGE_SIZE);
      setAppliedQuery(search);
      completedQuery.current = search;
    } catch (failure) {
      if (abort.signal.aborted || id !== requestId.current) return;
      const message = failure instanceof Error ? failure.message : 'Events could not be loaded. Please try again.';
      if (append) setPageError(message); else setError(message);
      setAppliedQuery(search);
      completedQuery.current = search;
    } finally {
      if (id === requestId.current) { setLoading(false); setLoadingMore(false); }
    }
  }, []);

  useEffect(() => {
    if (query.trim() === completedQuery.current) return;
    const timer = setTimeout(() => { void request(query.trim()); }, 350);
    searchTimer.current = timer;
    return () => { clearTimeout(timer); controller.current?.abort(); };
  }, [query, request]);
  useEffect(() => () => { controller.current?.abort(); }, []);

  const pending = loading || query.trim() !== appliedQuery;
  const now = Date.now();
  const visible = events.filter((event) => new Date(event.ends_at).getTime() > now)
    .sort((a, b) => new Date(a.starts_at).getTime() - new Date(b.starts_at).getTime());

  return <div className="feed-container discovery-page">
    <section className="discovery-heading" aria-labelledby="discover-heading">
      <div><h1 id="discover-heading">Discover events</h1><p>Find hackathons, meetups and conferences on Hackways. Explore a shared interest, meet a community or plan a gathering of your own.</p></div>
      <form id="event-search-form" className="discovery-search" role="search" onSubmit={(action) => { action.preventDefault(); clearTimeout(searchTimer.current); void request(query.trim()); }}>
        <Icon name="search" size={20} /><label className="visually-hidden" htmlFor="event-search">Search events by topic or keyword</label><input id="event-search" type="search" autoComplete="off" placeholder="Search events or topics" value={query} onChange={(action) => setQuery(action.target.value)} />
        {query && <button type="button" className="icon-button" aria-label="Clear search" onClick={() => setQuery('')}><Icon name="close" size={16} /></button>}
        <button type="submit" className="search-submit" aria-label="Search events"><Icon name="arrow" size={18} /></button>
      </form>
    </section>

    {!query.trim() && <DiscoverPromotions initial={promotions} />}
    <section className="feed-results" aria-labelledby="results-heading" aria-busy={pending || loadingMore}>
      <div className="feed-results-heading"><h2 id="results-heading">Upcoming gatherings</h2><span role="status" aria-live="polite">{pending ? 'Searching...' : error ? '' : `${visible.length} ${visible.length === 1 ? 'event' : 'events'}${hasMore ? ' loaded' : ''}`}</span></div>
      {appliedQuery && !pending && <p className="search-caption">Results for <strong>{appliedQuery}</strong><button type="button" className="quiet-link" onClick={() => setQuery('')}>Clear search <Icon name="close" size={14} /></button></p>}
      {pending ? <div className="feed-loading" role="status"><span className="visually-hidden">Finding events...</span><EventCardSkeleton /><EventCardSkeleton /></div>
        : error ? <div className="feed-state" role="alert"><Icon name="alert" size={25} /><h3>We couldn&apos;t load events.</h3><p>{error}</p><Button onClick={() => { void request(query.trim()); }}>Try again <Icon name="reload" size={16} /></Button><a href={`${site.public}/help`}>Get help <Icon name="external" size={14} /></a></div>
        : visible.length === 0 ? <div className="feed-state"><Icon name="calendar" size={26} /><h3>{query ? 'No events match your search.' : 'No upcoming events yet.'}</h3><p>{query ? 'Try another event name or topic.' : 'Check back later or host a gathering of your own.'}</p>{query ? <Button variant="outline" onClick={() => setQuery('')}>Clear search</Button> : <Link className="app-button app-button-outline" href="/create">Create an event <Icon name="arrow" size={16} /></Link>}</div>
        : <div className="catalog-list">{visible.map(event => <EventCard key={event.id} event={event} />)}</div>}
      {!pending && !error && hasMore && <div className="load-more">{pageError && <p role="alert">{pageError}</p>}<Button variant="outline" isLoading={loadingMore} onClick={() => { void request(appliedQuery, true); }}>{loadingMore ? 'Loading events...' : pageError ? 'Try loading more again' : 'Load more events'}</Button></div>}
    </section>
  </div>;
}
