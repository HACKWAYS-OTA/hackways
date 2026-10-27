import { AdminShell } from '@/components/admin/AdminShell';
import { SeoEditor } from '@/components/admin/SeoEditor';
import { UtilityNavigation } from '@/components/admin/UtilityNavigation';
import { adminServer } from '@/lib/admin-server';
import type { SeoTopic } from '@/lib/utilities';
import '../../utilities.css';
export default async function NewSeo() { const library=await adminServer<{data:SeoTopic[]}>('global','/utilities/seo/library').then(value=>value.data).catch(()=>null);return <AdminShell active="/utilities/seo-manager"><UtilityNavigation active="seo-manager" /><header className="utility-heading"><h1>Manage a page</h1></header><SeoEditor topics={library} /></AdminShell>; }
