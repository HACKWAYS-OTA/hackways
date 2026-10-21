import Link from 'next/link';
export function RegistrationTabs({ slug, questions = false }: { slug: string; questions?: boolean }) {
  return <nav className="host-page-tabs" aria-label="Registration settings"><Link href={`/events/${slug}/registration`} aria-current={!questions ? 'page' : undefined}>Joining & places</Link><Link href={`/events/${slug}/registration/questions`} aria-current={questions ? 'page' : undefined}>Registration questions</Link></nav>;
}
