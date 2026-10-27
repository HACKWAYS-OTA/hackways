import { AdminShell } from '@/components/admin/AdminShell';
import { SeoEditor } from '@/components/admin/SeoEditor';
import { UtilityNavigation } from '@/components/admin/UtilityNavigation';
import '../../utilities.css';
export default function NewSeo() { return <AdminShell active="/utilities/seo-manager"><UtilityNavigation active="seo-manager" /><header className="utility-heading"><h1>Manage a page</h1></header><SeoEditor /></AdminShell>; }
