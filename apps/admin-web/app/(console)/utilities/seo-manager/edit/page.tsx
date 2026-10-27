import { AdminShell } from '@/components/admin/AdminShell';
import { SeoEditor } from '@/components/admin/SeoEditor';
import { UtilityNavigation } from '@/components/admin/UtilityNavigation';
import { adminServer } from '@/lib/admin-server';
import type { SeoPage } from '@/lib/utilities';
import '../../utilities.css';
export default async function EditSeo({ searchParams }: { searchParams: { path?: string } }) {
 const path = (searchParams.path || '/').slice(0,240);
 const result = await adminServer<{ data: SeoPage | null }>('global', `/utilities/seo/page?path=${encodeURIComponent(path)}`);
 return <AdminShell active="/utilities/seo-manager"><UtilityNavigation active="seo-manager" /><header className="utility-heading"><h1>Page SEO</h1></header><SeoEditor key={path} initial={result.data || undefined} path={path} /></AdminShell>;
}
