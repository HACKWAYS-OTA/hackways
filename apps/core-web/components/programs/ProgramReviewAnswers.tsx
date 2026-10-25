import Link from 'next/link';
import type { Program, ProgramProfile } from '@hackways/api-client';

export function ProgramReviewAnswers({ program, values, email }: { program: Program; values: ProgramProfile; email: string }) {
  return <dl className="program-review-answers">
    <div><dt>Your details<Link href={`/programs/${program.slug}/apply/profile`}>Edit details</Link></dt><dd>{values.full_name}<br />{email}<br />{values.city}</dd></div>
    <div><dt>WhatsApp number</dt><dd>{values.phone || 'Not provided'}</dd></div>
    <div><dt>{program.slug === 'ambassadors' ? 'LinkedIn profile' : 'Profile link'}</dt><dd>{values.profile_url ? <a href={values.profile_url} target="_blank" rel="noopener noreferrer">{values.profile_url}</a> : 'Not provided'}</dd></div>
    {values.biography && <div><dt>Your introduction</dt><dd>{values.biography}</dd></div>}
    <div><dt>Your event idea<Link href={`/programs/${program.slug}/apply/proposal`}>Edit idea</Link></dt><dd>{values.proposal || 'No proposal added yet.'}</dd></div>
    {program.questions.map(question => <div key={question.id}><dt>{question.label}</dt><dd>{values.answers[question.id] || 'Not answered'}</dd></div>)}
  </dl>;
}
