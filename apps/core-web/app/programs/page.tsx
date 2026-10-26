import { pageSocialMetadata } from '@/lib/social-metadata';
import Link from 'next/link';
import type { Program } from '@hackways/api-client';
import { fetchApi } from '@/lib/api';
import { ProgramFrame } from '@/components/programs/ProgramFrame';
import { ProgramBanner } from '@/components/programs/ProgramBanner';
export const metadata = {
  ...pageSocialMetadata('programs'),
  title: 'Programs',
  description: 'Explore Hackways programs and open roles. Find opportunities to contribute to your community, host events and bring people together.',
  alternates: { canonical: '/programs' },
};
export default async function Programs() {
  const result = await fetchApi<{ data: Program[]; has_next: boolean }>('/programs', { next: { revalidate: 30 } });
  const now = Date.now();
  return <ProgramFrame title="Programs" section="Explore opportunities" directory>
    <p className="program-directory-intro">Bring your ideas, contribute to your community and find a program that fits.</p>
    <div className="program-directory">{result.data.map(program => {
      const open = program.status === 'open' && (!program.opens_at || new Date(program.opens_at).getTime() <= now) && (!program.closes_at || new Date(program.closes_at).getTime() > now);
      return <Link key={program.id} className={`program-directory-row${program.slug === 'ambassadors' ? ' program-directory-row-with-image' : ''}`} href={`/programs/${program.slug}`}>
        <div className="program-directory-copy">
          <span className="program-status">{open ? 'Applications open' : 'Applications closed'}</span>
          <h2>{program.title}</h2><p>{program.summary}</p><span className="program-directory-action">Explore the program</span>
        </div><ProgramBanner slug={program.slug} />
      </Link>;
    })}{result.data.length === 0 && <p>There are no published programs right now.</p>}</div>
  </ProgramFrame>;
}
