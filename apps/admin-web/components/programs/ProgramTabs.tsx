import Link from 'next/link';
export function ProgramTabs({ slug, active }: { slug: string; active: 'applications' | 'settings' | 'results' }) {
 return <nav className="program-admin-tabs" aria-label="Program workspace">{[{ id: 'applications', label: 'Applications', suffix: '' }, { id: 'settings', label: 'Program settings', suffix: '/settings' }, { id: 'results', label: 'Release results', suffix: '/results' }].map(tab => <Link key={tab.id} href={`/programs/${slug}${tab.suffix}`} aria-current={active === tab.id ? 'page' : undefined}>{tab.label}</Link>)}</nav>;
}
