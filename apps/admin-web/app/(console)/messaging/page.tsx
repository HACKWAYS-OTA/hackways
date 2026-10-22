import type { Metadata } from 'next';
import { AdminShell } from '@/components/admin/AdminShell';
import { Catalogue } from '@/components/admin/Catalogue';
import { catalogueParams, type CatalogueSearch } from '@/lib/catalogue-paths';
import { adminModules } from '@/lib/admin-modules';

export const metadata: Metadata = { title: `${adminModules.messaging.label} | Hackways Admin` };
export default function Page({ searchParams }: { searchParams: CatalogueSearch }) {
  const { query, page, status } = catalogueParams(searchParams);
  const filter = status === 'all' ? '' : status;
  return <AdminShell active="/messaging" query={query} status={filter}><Catalogue resource="messaging" query={query} page={page} status={filter} /></AdminShell>;
}
