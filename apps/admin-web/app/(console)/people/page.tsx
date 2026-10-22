import type { Metadata } from 'next';
import { AdminShell } from '@/components/admin/AdminShell';
import { Catalogue } from '@/components/admin/Catalogue';
import { catalogueParams, type CatalogueSearch } from '@/lib/catalogue-paths';
import { adminModules } from '@/lib/admin-modules';

export const metadata: Metadata = { title: `${adminModules.people.label} | Hackways Admin` };
export default function Page({ searchParams }: { searchParams: CatalogueSearch }) {
  const { query, page, status } = catalogueParams(searchParams);
  const filter = status === 'all' ? '' : status;
  return <AdminShell active="/people" query={query} status={filter}><Catalogue resource="people" query={query} page={page} status={filter} /></AdminShell>;
}
