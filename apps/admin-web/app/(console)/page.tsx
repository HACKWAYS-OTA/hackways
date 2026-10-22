import type { Metadata } from 'next';
import { AdminShell } from '@/components/admin/AdminShell';
import { Icon } from '@/components/ui/Icon';
import { adminModules, adminModuleOrder } from '@/lib/admin-modules';

export const metadata: Metadata = { title: 'Home | Hackways Admin' };
export default function AdminHomePage() {
  return <AdminShell active="/">
    <div className="admin-page-heading admin-home-heading"><h1>Welcome to Hackways!</h1></div>
    <nav className="admin-directory" aria-label="Administration areas">
      {adminModuleOrder.map(name => <a href={`/${name}`} key={name}><Icon name={name} size={26} /><div><h2>{adminModules[name].label}</h2><p>{adminModules[name].description}</p></div><Icon name="arrow" /></a>)}
    </nav>
  </AdminShell>;
}
