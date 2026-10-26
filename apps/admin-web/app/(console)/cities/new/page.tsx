import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { CityEditor } from '@/components/admin/CityEditor';
import '../cities.css';
export default function NewCity() { return <AdminShell active="/cities"><Link className="quiet-link" href="/cities">Back to City Hubs</Link><header className="city-admin-heading"><div><h1>Create a City Hub</h1><p className="admin-secondary-copy">Prepare a city page, add its banner and curate public events.</p></div></header><CityEditor /></AdminShell>; }
