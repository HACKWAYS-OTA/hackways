'use client';

import Link from 'next/link';
import { useEffect, useRef, useState } from 'react';
import { AnalyticsRangeControl } from './AnalyticsRangeControl';
import { adminRequest } from '@/lib/admin-request';
import { analyticsNumber as number, type LiveCoreAnalytics, type LiveGlobalAnalytics, type AnalyticsWindow } from '@/lib/analytics';

type Reports = { core: LiveCoreAnalytics | null; global: LiveGlobalAnalytics | null };

function Activity({ rows }: { rows: LiveCoreAnalytics['activity'] }) {
  const max = Math.max(1, ...rows.flatMap(row => [row.registrations, row.check_ins]));
  const step = 600 / Math.max(1, rows.length);
  return <section className="analytics-panel analytics-chart"><header><h2>Registration activity</h2></header>
    {rows.some(row => row.registrations || row.check_ins) ? <>
      <svg viewBox="0 0 640 200" role="img" aria-label="Registrations and check-ins, from the start of the selected interval to now">
        {[0, max].map(value => <text key={value} x="0" y={175 - value / max * 150}>{number(value)}</text>)}
        {rows.map((row, index) => <g key={row.bucket}><title>{`Period ${index + 1}: ${number(row.registrations)} registrations, ${number(row.check_ins)} check-ins`}</title>
          <rect className="analytics-bar-primary" x={32 + index * step} y={170 - row.registrations / max * 150} width={step * .36} height={row.registrations / max * 150} />
          <rect className="analytics-bar-secondary" x={32 + index * step + step * .4} y={170 - row.check_ins / max * 150} width={step * .36} height={row.check_ins / max * 150} />
        </g>)}
        <text x="32" y="197">Start</text><text x="632" y="197" textAnchor="end">Now</text>
      </svg>
      <ul className="analytics-legend"><li><span className="is-primary" />Registrations</li><li><span className="is-secondary" />Check-ins</li></ul>
      <details className="analytics-daily-data"><summary>View values</summary><div className="analytics-table-scroll" tabIndex={0} role="region" aria-label="Registration activity values"><table><thead><tr><th scope="col">Period</th><th scope="col">Registrations</th><th scope="col">Check-ins</th></tr></thead><tbody>{rows.map((row, index) => <tr key={row.bucket}><th scope="row">{index + 1}</th><td>{number(row.registrations)}</td><td>{number(row.check_ins)}</td></tr>)}</tbody></table></div></details>
    </> : <p className="analytics-chart-empty">No registrations or check-ins in this interval.</p>}
  </section>;
}

function Breakdown({ title, rows }: { title: string; rows: { label: string; value: number }[] }) {
  return <section className="analytics-panel"><h2>{title}</h2>{rows.length ? <dl className="analytics-breakdown">{rows.map(row => <div key={row.label}><dt>{row.label}</dt><dd>{number(row.value)}</dd></div>)}</dl> : <p className="analytics-empty">No activity in this interval.</p>}</section>;
}

export function LiveAnalytics({ initial, initialWindow, today, from }: { initial: Reports; initialWindow: AnalyticsWindow; today: string; from: string }) {
  const [window, setWindow] = useState(initialWindow);
  const [reports, setReports] = useState(initial);
  const [loading, setLoading] = useState(false);
  const [retry, setRetry] = useState(0);
  const firstMount = useRef(true);
  useEffect(() => {
    let active = true;
    let timer: ReturnType<typeof setTimeout> | undefined;
    let request: AbortController | undefined;
    const refresh = async () => {
      if (!active || document.hidden || request) return;
      const controller = new AbortController();
      request = controller;
      setLoading(true);
      const [core, global] = await Promise.allSettled([
        adminRequest<LiveCoreAnalytics>(`core/analytics/live?window=${window}`, { signal: controller.signal }),
        adminRequest<LiveGlobalAnalytics>(`global/analytics/live?window=${window}`, { signal: controller.signal }),
      ]);
      if (!active || controller.signal.aborted) return;
      setReports({ core: core.status === 'fulfilled' && core.value.window === window ? core.value : null, global: global.status === 'fulfilled' && global.value.window === window ? global.value : null });
      setLoading(false);
      request = undefined;
      timer = setTimeout(refresh, 15000);
    };
    const visibility = () => {
      clearTimeout(timer);
      request?.abort();
      request = undefined;
      if (!document.hidden) void refresh();
      else setLoading(false);
    };
    document.addEventListener('visibilitychange', visibility);
    if (firstMount.current && initial.core && initial.global) timer = setTimeout(refresh, 15000);
    else void refresh();
    firstMount.current = false;
    return () => { active = false; clearTimeout(timer); request?.abort(); document.removeEventListener('visibilitychange', visibility); };
  }, [window, retry, initial, initialWindow]);
  const { core, global } = reports;
  return <>
    <header className="analytics-heading analytics-live-heading"><h1>Analytics</h1><div className="analytics-live-controls">
      <AnalyticsRangeControl value={window} from={from} to={today} today={today} onIntervalChange={value => {
        if (value === window) return;
        setReports({ core: null, global: null }); setLoading(true); setWindow(value);
        const url = new URL(location.href); url.search = new URLSearchParams({ window: value }).toString(); history.replaceState(history.state, '', url);
      }} />
      <button type="button" className="app-button app-button-outline app-button-sm" disabled={loading} onClick={() => setRetry(value => value + 1)}>{loading ? 'Refreshing...' : 'Refresh'}</button>
    </div></header>
    {!loading && !core && <p className="journey-error" role="alert">Event analytics could not be loaded. Refresh to retry.</p>}
    {!loading && !global && <p className="journey-error" role="alert">Visitor and account analytics could not be loaded. Refresh to retry.</p>}
    {loading && !core && !global && <div className="admin-result-state admin-loading" role="status" aria-label="Loading analytics"><span className="wordmark admin-loading-wordmark" aria-hidden="true">hackways.</span><span className="admin-loading-line" aria-hidden="true" /></div>}
    {global?.limited && <p className="journey-error" role="status">Visitor collection reached its limit. Visitor counts are partial.</p>}
    <div aria-busy={loading}>
      <dl className="analytics-summary">
        {core && <><div><dt>Registrations</dt><dd>{number(core.summary.registrations)}</dd></div><div><dt>Check-ins</dt><dd>{number(core.events.check_ins)}</dd></div><div><dt>Events created</dt><dd>{number(core.events.events_created)}</dd></div></>}
        {global && <><div><dt>New accounts</dt><dd>{number(global.accounts.new_accounts)}</dd></div><div><dt>Consented visitors</dt><dd>{number(global.traffic.visitors)}</dd></div><div><dt>Page views</dt><dd>{number(global.traffic.page_views)}</dd></div></>}
      </dl>
      <div className="analytics-main-grid">{core && <Activity rows={core.activity} />}{core && <Breakdown title="Registration states" rows={[{ label: 'Confirmed', value: core.summary.confirmed }, { label: 'Pending approval', value: core.summary.pending }, { label: 'Waiting / offered', value: core.summary.waiting }, { label: 'Other outcomes', value: core.summary.other }]} />}</div>
      {core && <section className="analytics-panel"><header><h2>Events by registrations</h2><Link className="analytics-history-link" href="/events">All events</Link></header>{core.top_events.length ? <dl className="analytics-breakdown">{core.top_events.map(event => <div key={event.reference}><dt><Link href={`/events/details/${event.reference}`}>{event.title}</Link></dt><dd>{number(event.registrations)}</dd></div>)}</dl> : <p className="analytics-empty">No registrations in this interval.</p>}</section>}
      {global && <><div className="analytics-main-grid"><Breakdown title="Top pages" rows={global.pages.map(row => ({ label: row.route, value: row.views }))} /><Breakdown title="Sources" rows={global.sources.map(row => ({ label: row.source, value: row.views }))} /></div><section className="analytics-panel"><header><h2>Messaging</h2><Link href="/messaging" className="analytics-history-link">Delivery queue</Link></header><dl className="analytics-breakdown"><div><dt>Emails accepted by provider</dt><dd>{number(global.accounts.accepted_emails)}</dd></div><div><dt>Failed emails created in this interval</dt><dd>{number(global.accounts.failed_emails)}</dd></div></dl></section></>}
    </div>
  </>;
}
