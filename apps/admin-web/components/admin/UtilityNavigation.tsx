import Link from 'next/link';
export function UtilityNavigation({ active }: { active: 'notice' | 'seo-manager' }) {
 return <nav className="utility-navigation" aria-label="Utilities"><Link href="/utilities/notice" aria-current={active === 'notice' ? 'page' : undefined}>Notices</Link><Link href="/utilities/seo-manager" aria-current={active === 'seo-manager' ? 'page' : undefined}>SEO Manager</Link></nav>;
}
