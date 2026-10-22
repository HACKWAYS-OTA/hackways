import type { Metadata } from 'next';
import { notFound, redirect } from 'next/navigation';
import { AdminShell } from '@/components/admin/AdminShell';
import { AdminDetail } from '@/components/admin/AdminDetail';
import { Icon } from '@/components/ui/Icon';
import { getAdminDetail } from '@/lib/admin-details';
import { isAdminModule, relatedPath } from '@/lib/admin-detail-paths';
import { adminModules } from '@/lib/admin-modules';

type Props = { params: { resource: string; id: string }; searchParams: { section?: string; offset?: string; kind?: string } };
export function generateMetadata({ params }: Props): Metadata {
  return { title: `${isAdminModule(params.resource) ? adminModules[params.resource].label : 'Record'} details | Hackways Admin` };
}

export default async function Page({ params, searchParams }: Props) {
  if (!isAdminModule(params.resource)) notFound();
  const resource = params.resource;
  const id = params.id;
  const numericReference = /^[1-9][0-9]*$/.test(id);
  const legacyID = resource !== 'audit' && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id);
  if (!numericReference && !legacyID) notFound();
  const section = typeof searchParams.section === 'string' ? searchParams.section : '';
  const offsetValue = Number(searchParams.offset || '0');
  const offset = Number.isSafeInteger(offsetValue) && offsetValue >= 0 && offsetValue <= 1000000 ? offsetValue : 0;
  const kind = resource === 'storage' && typeof searchParams.kind === 'string' ? searchParams.kind : '';
  const detail = await getAdminDetail(resource, id, section, offset, kind);
  if (detail && legacyID) redirect(relatedPath(resource, detail.data.reference, detail.related.section, detail.related.offset, kind));
  return <AdminShell active={`/${resource}`}>
    {detail ? <AdminDetail resource={resource} detail={detail} /> : <div className="admin-result-state" role="alert"><p>This record could not be loaded.</p><a className="app-button app-button-outline" href={relatedPath(resource, id, '', 0, kind)}>Try again</a><a className="app-button app-button-outline app-button-sm" href={`/${resource}`}><Icon name="back" size={17} />Back</a></div>}
  </AdminShell>;
}
