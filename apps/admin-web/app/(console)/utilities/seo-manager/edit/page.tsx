import { AdminShell } from '@/components/admin/AdminShell';
import { SeoEditor } from '@/components/admin/SeoEditor';
import { UtilityNavigation } from '@/components/admin/UtilityNavigation';
import { adminServer } from '@/lib/admin-server';
import type { SeoPage, SeoTopic } from '@/lib/utilities';
import '../../utilities.css';
export default async function EditSeo({ searchParams }: { searchParams: { path?: string } }) {
 const path = (searchParams.path || '/').slice(0,240);
 const [result,library] = await Promise.all([adminServer<{ data: SeoPage | null }>('global', `/utilities/seo/page?path=${encodeURIComponent(path)}`),adminServer<{data:SeoTopic[]}>('global','/utilities/seo/library').then(value=>value.data).catch(()=>null)]);
 return <AdminShell active="/utilities/seo-manager"><UtilityNavigation active="seo-manager" /><header className="utility-heading"><h1>Page SEO</h1></header><SeoEditor key={path} initial={result.data || undefined} path={path} topics={library} /></AdminShell>;
}
