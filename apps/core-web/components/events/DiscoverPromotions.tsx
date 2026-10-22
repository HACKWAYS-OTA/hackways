'use client';

import { useEffect, useState } from 'react';
import { EventCard } from './EventCard';
import { activePromotions, type PromotionFeed } from '@/lib/promotions';

export function DiscoverPromotions({ initial }: { initial: PromotionFeed | null }) {
  const [feed, setFeed] = useState(initial);
  const [index, setIndex] = useState(0);
  const [paused, setPaused] = useState(false);
  const [interacting, setInteracting] = useState(false);
  const [hidden, setHidden] = useState(false);
  const [reduced, setReduced] = useState(false);
  const [explicitPlay, setExplicitPlay] = useState(false);
  const [clock, setClock] = useState(() => Date.parse(initial?.server_time || '') || Date.now());
  useEffect(() => {
    const preference = window.matchMedia('(prefers-reduced-motion: reduce)');
    const motion = () => setReduced(preference.matches);
    motion(); preference.addEventListener('change', motion);
    return () => preference.removeEventListener('change', motion);
  }, []);
  useEffect(() => {
    let disposed = false, request: AbortController | null = null, timer: ReturnType<typeof setTimeout>;
    let offset = feed ? Date.parse(feed.server_time) - Date.now() : 0;
    async function refresh() {
      request?.abort(); request = new AbortController();
      try {
        const next = await activePromotions(request.signal);
        if (disposed) return;
        offset = Date.parse(next.server_time) - Date.now();
        setFeed(next); setClock(Date.now() + offset);
        schedule(next.next_change_at);
      } catch {
        if (disposed || request.signal.aborted) return;
        // Hide this optional placement on an unavailable feed; never substitute
        // a catalogue event or retain an expired promotion.
        setFeed(null); schedule(null);
      }
    }
    function schedule(boundary: string | null) {
      clearTimeout(timer);
      if (document.hidden) return;
      const wait = boundary ? Math.max(1000, Date.parse(boundary) - (Date.now() + offset) + 150) : 60000;
      timer = setTimeout(() => { setClock(Date.now() + offset); void refresh(); }, Math.min(60000, wait));
    }
    function visibility() { setHidden(document.hidden); if (document.hidden) { clearTimeout(timer); request?.abort(); } else { setClock(Date.now() + offset); void refresh(); } }
    setHidden(document.hidden); schedule(feed?.next_change_at || null);
    document.addEventListener('visibilitychange', visibility);
    return () => { disposed = true; clearTimeout(timer); request?.abort(); document.removeEventListener('visibilitychange', visibility); };
    // The refresh loop owns subsequent feed values and boundary timers.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  const active = feed?.data.filter(item => Date.parse(item.visible_until) > clock) || [];
  const count = active.length;
  const playing = !paused && (!reduced || explicitPlay);
  useEffect(() => { if (count < 2 || !playing || interacting || hidden) return; const timer = setInterval(() => setIndex(value => (value + 1) % count), 7000); return () => clearInterval(timer); }, [count, playing, interacting, hidden]);
  if (!count) return null;
  const current = index % count;
  return <section className="discover-promotions" aria-label="Promoted events" aria-roledescription="carousel" onMouseEnter={() => setInteracting(true)} onMouseLeave={() => setInteracting(false)} onFocusCapture={() => setPaused(true)}>
    <div role="group" aria-roledescription="slide" aria-label={`${current + 1} of ${count}`}><EventCard event={active[current].events} prominent /></div>
    {count > 1 && <div className="promotion-navigation"><div className="promotion-pages" aria-label="Choose promoted event">{active.map((item, position) => <button key={item.slot} type="button" aria-label={`Show ${item.events.title}`} aria-current={current === position ? 'true' : undefined} onClick={() => { setPaused(true); setIndex(position); }}>{position + 1}</button>)}</div><button type="button" className="app-button app-button-outline app-button-sm" onClick={() => { setPaused(playing); setExplicitPlay(true); }}>{playing ? 'Pause slides' : 'Play slides'}</button></div>}
  </section>;
}
