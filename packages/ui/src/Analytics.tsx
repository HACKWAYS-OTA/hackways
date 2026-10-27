'use client';

import { useEffect, useRef, useState } from 'react';

const consentCookie = 'hackways_analytics_consent';
const visitorCookie = 'hackways_analytics_visitor';
const sessionCookie = 'hackways_analytics_session';
const productEvent = 'hackways-analytics-event';
const preferencesEvent = 'hackways-cookie-preferences';
const staticPaths = new Set(['/', '/discover', '/events', '/spaces', '/spaces/new', '/cities', '/programs', '/create', '/my-events', '/account', '/login', '/not-allowed', '/about', '/contact', '/acceptable-use', '/cookies', '/terms', '/privacy', '/help', '/organizers', '/ambassadors', '/mcp', '/mcp/setup', '/connect/action', '/connect/authorize', '/connect/connections', '/manage', '/manage/spaces', '/manage/onboarding', '/manage/create', '/manage/login']);
const authPath = /^\/(?:manage\/)?auth\/(?:forgot-password|reset-password|verify-email|google\/callback|unsubscribe)$/;
const publicEntityPath = /^\/(?:events\/[a-zA-Z0-9-]{1,180}(?:\/register)?|(?:spaces|cities|programs|help)\/[a-zA-Z0-9-]{1,180})$/;

/** Canonical page names only. Private identifiers and query strings never enter a visit. */
function analyticsRoute(path: string): string | null {
  if (!path.startsWith('/') || /[?#%\\]/.test(path)) return null;
  const route = path.length > 1 ? path.replace(/\/$/, '') : path;
  if (staticPaths.has(route) || authPath.test(route)) return route;
  if (/^\/my-events\/[^/]+$/.test(route)) return '/my-events/:registration';
  if (/^\/spaces\/[^/]+\/settings$/.test(route)) return '/spaces/:space/settings';
  if (/^\/programs\/[^/]+\/application$/.test(route)) return '/programs/:program/application';
  if (/^\/programs\/[^/]+\/apply\/[^/]+$/.test(route)) return '/programs/:program/apply/:step';
  const workspace = route.match(/^\/manage\/(events|spaces)\/[^/]+(?:\/(team|settings|registration(?:\/questions)?|check-in|invitations|appearance|edit|guests|activity))?$/);
  if (workspace && (workspace[1] === 'events' || !workspace[2] || workspace[2] === 'team')) return `/manage/${workspace[1]}/:${workspace[1] === 'events' ? 'event' : 'space'}${workspace[2] ? `/${workspace[2]}` : ''}`;
  return publicEntityPath.test(route) ? route : null;
}
type ProductEvent = 'registration_step' | 'registration_complete';
type Visit = { id: string; event_name: 'page_view' | ProductEvent; route: string; device: 'mobile' | 'tablet' | 'desktop'; referrer_host: string; step?: number };

function cookie(name: string): string {
  const item = document.cookie.split(';').map(value => value.trim()).find(value => value.startsWith(`${name}=`));
  try { return item ? decodeURIComponent(item.slice(name.length + 1)) : ''; } catch { return ''; }
}

function saveCookie(name: string, value: string, seconds: number) {
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${seconds}; SameSite=Lax${window.location.protocol === 'https:' ? '; Secure' : ''}`;
}

function identity(name: string, seconds: number): string | null {
  let value = cookie(name);
  if (!/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)) {
    if (!window.crypto.randomUUID) return null;
    value = window.crypto.randomUUID();
  }
  saveCookie(name, value, seconds);
  return cookie(name) === value ? value : null;
}

/** Sends only a named product event; answers, account IDs and full URLs are never accepted. */
export function recordAnalyticsEvent(name: ProductEvent, step?: number) {
  if (typeof window !== 'undefined') window.dispatchEvent(new CustomEvent(productEvent, { detail: { name, step } }));
}

export function CookiePreferencesButton({ label = 'Cookie preferences' }: { label?: string } = {}) {
  return <button type="button" className="hw-cookie-preferences" onClick={() => window.dispatchEvent(new Event(preferencesEvent))}>{label}</button>;
}

export function AnalyticsTracker({ path, endpoint = '/api/v1/analytics/events' }: { path: string; endpoint?: string }) {
  const [choice, setChoice] = useState<string | null>(null), [open, setOpen] = useState(false), [error, setError] = useState('');
  const route = analyticsRoute(path);
  const currentPath = useRef(route), currentPage = useRef(path), lastPage = useRef('');
  const queueVisit = useRef<(name: Visit['event_name'], step?: number) => void>(() => undefined);
  const stopTracking = useRef<() => void>(() => undefined), heading = useRef<HTMLHeadingElement>(null);
  currentPath.current = route;
  currentPage.current = path;

  useEffect(() => {
    const stored = cookie(consentCookie); setChoice(stored);
    if (!stored && currentPath.current) setOpen(true);
    const show = () => { setOpen(true); requestAnimationFrame(() => heading.current?.focus({ preventScroll: true })); };
    window.addEventListener(preferencesEvent, show);
    return () => window.removeEventListener(preferencesEvent, show);
  }, []);

  useEffect(() => {
    if (choice !== 'yes') return;
    let queue: Visit[] = [], busy = false, timer: ReturnType<typeof setTimeout> | undefined, pausedUntil = 0;
    let request: AbortController | undefined;
    const referrer = (() => {
      try {
        const host = new URL(document.referrer).hostname.toLowerCase();
        return host === window.location.hostname || host === 'hackways.in' || host.endsWith('.hackways.in') ? '' : /^[a-z0-9.-]{1,253}$/.test(host) ? host : '';
      } catch { return ''; }
    })();
    async function flush(retry = false) {
      if (cookie(consentCookie) !== 'yes') { queue = []; return; }
      if (busy || !queue.length || Date.now() < pausedUntil) return;
      const visitor = identity(visitorCookie, 90 * 86400), session = identity(sessionCookie, 30 * 60);
      if (!visitor || !session) { queue = []; return; }
      busy = true; const batch = queue.splice(0, 10); request = new AbortController();
      const timeout = setTimeout(() => request?.abort(), 10000);
      try {
        const response = await fetch(endpoint, { method: 'POST', credentials: 'include', keepalive: true, signal: request.signal,
          headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ consent: true, visitor_id: visitor, session_id: session, events: batch }) });
        if (response.status === 429) { pausedUntil = Date.now() + 3600000; queue = []; }
        else if (response.status >= 500 && !retry && cookie(consentCookie) === 'yes') { queue = [...batch, ...queue].slice(0, 20); timer = setTimeout(() => { timer = undefined; void flush(true); }, 2000); }
      } catch {
        if (!retry && !request.signal.aborted && cookie(consentCookie) === 'yes') { queue = [...batch, ...queue].slice(0, 20); timer = setTimeout(() => { timer = undefined; void flush(true); }, 2000); }
      } finally { clearTimeout(timeout); busy = false; if (queue.length && !timer) timer = setTimeout(() => { timer = undefined; void flush(true); }, 1000); }
    }
    const schedule = (name: Visit['event_name'], step?: number) => {
      const route = currentPath.current;
      if (cookie(consentCookie) !== 'yes' || !route || !window.crypto.randomUUID || Date.now() < pausedUntil) return;
      if (name !== 'page_view' && !route.endsWith('/register')) return;
      const device = window.matchMedia('(max-width:640px)').matches ? 'mobile' : window.matchMedia('(max-width:1024px)').matches ? 'tablet' : 'desktop';
      queue.push({ id: window.crypto.randomUUID(), event_name: name, route, device, referrer_host: referrer, ...(name === 'registration_step' && step !== undefined ? { step } : {}) });
      if (queue.length > 20) queue = queue.slice(-20);
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => { timer = undefined; void flush(); }, 600);
    };
    const custom = (action: Event) => {
      if (!(action instanceof CustomEvent)) return;
      const value: unknown = action.detail;
      if (!value || typeof value !== 'object' || !('name' in value)) return;
      if (value.name === 'registration_complete') schedule(value.name);
      else if (value.name === 'registration_step' && 'step' in value && typeof value.step === 'number' && Number.isInteger(value.step) && value.step >= 0 && value.step <= 50) schedule(value.name, value.step);
    };
    const hide = () => { if (document.visibilityState === 'hidden') { if (timer) clearTimeout(timer); timer = undefined; void flush(); } };
    queueVisit.current = schedule;
    stopTracking.current = () => { queue = []; request?.abort(); if (timer) clearTimeout(timer); };
    window.addEventListener(productEvent, custom); document.addEventListener('visibilitychange', hide);
    if (currentPath.current && lastPage.current !== currentPage.current) { lastPage.current = currentPage.current; schedule('page_view'); }
    return () => {
      window.removeEventListener(productEvent, custom); document.removeEventListener('visibilitychange', hide);
      if (timer) clearTimeout(timer); timer = undefined; void flush(); queueVisit.current = () => undefined;
    };
  }, [choice, endpoint]);

  useEffect(() => {
    if (choice === 'yes' && route && lastPage.current !== path) { lastPage.current = path; queueVisit.current('page_view'); }
    if (choice === '' && route) setOpen(true);
  }, [path, route, choice]);

  function choose(analytics: boolean) {
    const next = analytics ? 'yes' : 'no'; saveCookie(consentCookie, next, 365 * 86400);
    if (!analytics) { stopTracking.current(); saveCookie(visitorCookie, '', 0); saveCookie(sessionCookie, '', 0); lastPage.current = ''; }
    if (analytics && cookie(consentCookie) !== 'yes') { setError('Your browser could not save this choice. Analytics stays off.'); setChoice('no'); return; }
    setChoice(next); setOpen(false); setError('');
  }

  if (!open) return null;
  return <section className="hw-analytics-consent" aria-labelledby="hw-analytics-heading">
    <div><h2 id="hw-analytics-heading" ref={heading} tabIndex={-1}>Your cookie choices</h2><p>Essential cookies keep you signed in. Optional analytics cookies help us understand visits and improve Hackways.</p><a href="/privacy">Privacy policy</a>{error && <p role="alert">{error}</p>}</div>
    <div className="hw-analytics-choices"><button type="button" onClick={() => choose(true)}>Allow all</button><button type="button" onClick={() => choose(false)}>Essential only</button></div>
  </section>;
}
