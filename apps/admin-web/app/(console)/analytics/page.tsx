import { AdminShell } from '@/components/admin/AdminShell';
import { HistoricalAnalytics } from '@/components/admin/HistoricalAnalytics';
import { LiveAnalytics } from '@/components/admin/LiveAnalytics';
import { adminServer } from '@/lib/admin-server';
import type { AnalyticsWindow, LiveCoreAnalytics, LiveGlobalAnalytics } from '@/lib/analytics';
import '@/components/admin/analytics.css';

export const metadata = { title: 'Analytics | Hackways Admin' };

export default async function AnalyticsPage({ searchParams }: { searchParams: { view?: string; window?: string; from?: string; to?: string } }) {
  if (searchParams.view === 'history' || searchParams.from || searchParams.to) return <AdminShell active="/analytics"><HistoricalAnalytics searchParams={searchParams} /></AdminShell>;
  const window: AnalyticsWindow = searchParams.window === '1h' || searchParams.window === '24h' ? searchParams.window : '5m';
  const [core, global] = await Promise.allSettled([
    adminServer<LiveCoreAnalytics>('core', `/analytics/live?window=${window}`),
    adminServer<LiveGlobalAnalytics>('global', `/analytics/live?window=${window}`),
  ]);
  return <AdminShell active="/analytics"><LiveAnalytics key={window} initialWindow={window} initial={{ core: core.status === 'fulfilled' ? core.value : null, global: global.status === 'fulfilled' ? global.value : null }} /></AdminShell>;
}
