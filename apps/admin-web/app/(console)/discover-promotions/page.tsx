import { AdminShell } from '@/components/admin/AdminShell';
import { DiscoverPromotionEditor } from '@/components/admin/DiscoverPromotionEditor';
import { adminServer } from '@/lib/admin-server';
import type { PromotionConfiguration } from '@/lib/discover-promotions';
export const metadata = { title: 'Discover promotions | Hackways Admin' };
export default async function DiscoverPromotions() {
  const result = await adminServer<{ data: PromotionConfiguration }>('core', '/discover-promotions');
  return <AdminShell active="/discover-promotions"><div className="admin-page-heading"><h1>Discover promotions</h1></div><DiscoverPromotionEditor initial={result.data} /></AdminShell>;
}
