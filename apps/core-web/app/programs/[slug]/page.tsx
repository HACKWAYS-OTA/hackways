import Link from 'next/link';
import { programRead } from '@/lib/programs';
import { ProgramFrame } from '@/components/programs/ProgramFrame';
import { entitySocialMetadata } from '@/lib/social-metadata';
export async function generateMetadata({ params }: { params: { slug: string } }) {
  const { data } = await programRead(params.slug);
  return { ...entitySocialMetadata(`/programs/${encodeURIComponent(data.slug)}`, data.title, data.summary), title: data.title, description: data.summary };
}
export default async function ProgramOverview({ params }: { params: { slug: string } }) {
  const { data: program, accepting_applications: accepting } = await programRead(params.slug);
  const ambassador = program.slug === 'ambassadors';
  return <ProgramFrame title={program.title} slug={program.slug} section="Program details">
    <div className="program-detail-copy"><p className="program-lead">{program.summary}</p><p className="program-prose">{program.overview}</p></div>
    <div className="program-overview-grid"><article>
      <section><h2>Bring your community together.</h2><p className="program-prose">{program.responsibilities}</p>{ambassador && <p>Make space for people who share an interest to meet, learn and build connections. From a local meetup to a hands-on workshop, bring an idea that matters to your community and help turn it into a gathering worth showing up for.</p>}</section>
      {ambassador && <section><h2>There is room for your ideas.</h2><p>This is an open ambassador role for people who want to take the initiative and bring others along. You do not need a huge audience to begin. A thoughtful idea, a welcoming approach and the willingness to host are a place to start.</p></section>}
      <section><h2>Recognition for your work</h2><p className="program-prose">{program.recognition}</p></section>
      <section><h2>Start with your story.</h2><p>Tell us about yourself, your community and the event you would love to host. We will review your profile and share updates by email.{ambassador && ' Results will be shared at the end of November.'}</p></section>
    </article><aside className="program-apply-panel"><h2>Your next step.</h2><dl><div><dt>Applications</dt><dd>{accepting ? 'Open' : 'Currently closed'}</dd></div>{program.closes_at && <div><dt>Closes</dt><dd>{new Date(program.closes_at).toLocaleString('en-IN', { timeZone: 'Asia/Kolkata' })} IST</dd></div>}</dl>
      {accepting && <Link className="program-button" href={`/programs/${program.slug}/apply/profile`}>Apply for this role</Link>}<Link className="program-text-link" href={`/programs/${program.slug}/application`}>Already applied? View your application</Link><p>Your next gathering could start here.</p>
    </aside></div>
  </ProgramFrame>;
}
