import Link from 'next/link';
import { AnalyticsFilters } from '@/components/admin/AnalyticsFilters';
import { AnalyticsReport } from '@/components/admin/AnalyticsReport';
import { adminServer } from '@/lib/admin-server';
import { analyticsPeriod, type CoreAnalytics, type GlobalAnalytics } from '@/lib/analytics';


export async function HistoricalAnalytics({ searchParams }: { searchParams: { from?: string; to?: string } }) {
  const period = analyticsPeriod(searchParams);
  const [core, global] = period.valid ? await Promise.allSettled([
    adminServer<CoreAnalytics>('core', `/analytics?${period.query}`),
    adminServer<GlobalAnalytics>('global', `/analytics?${period.query}`),
  ]) : [];
  const events = core?.status === 'fulfilled' ? core.value : null;
  const accounts = global?.status === 'fulfilled' ? global.value : null;
  return <>
    <header className="analytics-heading"><h1>Analytics history</h1><div><Link className="app-button app-button-outline app-button-sm" href="/analytics">Live analytics</Link>{events && <a className="app-button app-button-outline app-button-sm" href={`/api/v1/admin/core/analytics?${period.query}&format=csv`}>Export events</a>}{accounts && <a className="app-button app-button-outline app-button-sm" href={`/api/v1/admin/global/analytics?${period.query}&format=csv`}>Export traffic & accounts</a>}</div></header>
    <AnalyticsFilters key={period.query} from={period.from} to={period.to} today={period.today} />
    {!period.valid ? <p className="journey-error" role="alert">Choose up to 90 days within the latest 90 days.</p> : <>{!events && <p className="journey-error" role="alert">Event analytics could not be loaded. Apply the date range to retry.</p>}{!accounts && <p className="journey-error" role="alert">Visitor and account analytics could not be loaded. Apply the date range to retry.</p>}<AnalyticsReport core={events} global={accounts} /></>}
  </>;
}
