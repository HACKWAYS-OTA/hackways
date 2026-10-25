'use client';
import { useEffect, useRef, useState, type FormEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Checkbox } from '@hackways/ui/checkbox';
import { Select } from '@hackways/ui/select';
import type { ProgramMine, ProgramProfile } from '@hackways/api-client';
import type { User } from '@/types/api';
import { fetchApi, ApiClientError } from '@/lib/api';
import { ProgramProfileFields } from './ProgramProfileFields';
import { ProgramReviewAnswers } from './ProgramReviewAnswers';

const steps = [{ id: 'profile', label: 'Your profile' }, { id: 'proposal', label: 'Your idea' }, { id: 'review', label: 'Review & submit' }] as const;
export function ProgramApplicationForm({ data, user, step }: { data: ProgramMine; user: User; step: typeof steps[number]['id'] }) {
 const router = useRouter(), program = data.program, existing = data.application, index = steps.findIndex(item => item.id === step);
 const [values, setValues] = useState<ProgramProfile>({ full_name: existing?.full_name || user.display_name, city: existing?.city || '', phone: existing?.phone || '', biography: existing?.biography || '', profile_url: existing?.profile_url || '', proposal: existing?.proposal || '', age: existing?.age || 0, age_confirmed: existing?.age_confirmed || false, answers: existing?.answers || {} });
 const revision = useRef(existing?.revision || 0);
 const [busy, setBusy] = useState(false), [error, setError] = useState(''), [notice, setNotice] = useState(''), [dirty, setDirty] = useState(false), [confirmed, setConfirmed] = useState(false), [needsSignIn, setNeedsSignIn] = useState(false);
 useEffect(() => { const warn = (event: BeforeUnloadEvent) => { if (dirty) { event.preventDefault(); event.returnValue = ''; } }; window.addEventListener('beforeunload', warn); return () => window.removeEventListener('beforeunload', warn); }, [dirty]);
 function change<K extends keyof ProgramProfile>(key: K, value: ProgramProfile[K]) { setValues(previous => ({ ...previous, [key]: value })); setDirty(true); setNotice(''); }
 async function save(destination?: string) {
  if (busy) return; setBusy(true); setError(''); setNotice(''); setNeedsSignIn(false);
  try {
   const answers = Object.fromEntries(program.questions.map(question => [question.id, values.answers[question.id] || '']));
   const saved = await fetchApi<ProgramMine>(`/programs/${program.slug}/application`, { method: 'POST', body: JSON.stringify({ ...values, answers, revision: revision.current, program_revision: program.revision }) });
   if (!saved.application) throw new Error('The saved application could not be confirmed. Please reload before trying again.');
   revision.current = saved.application.revision; setDirty(false); setNotice('Draft saved.');
   if (destination) router.push(destination);
  } catch (cause) { setNeedsSignIn(cause instanceof ApiClientError && cause.code === 'UNAUTHORIZED'); setError(cause instanceof Error ? cause.message : 'Your draft could not be saved. Your answers are still here.'); }
  finally { setBusy(false); }
 }
 async function navigate(destination: string) { if (busy) return; if (dirty) await save(destination); else router.push(destination); }
 async function submit(event: FormEvent<HTMLFormElement>) {
  event.preventDefault(); if (busy) return;
  if (step !== 'review') { await save(`/programs/${program.slug}/apply/${steps[index + 1].id}`); return; }
  if (!confirmed) { setError('Review your answers and confirm before submitting.'); return; }
  setBusy(true); setError(''); setNeedsSignIn(false);
  try { await fetchApi<ProgramMine>(`/programs/${program.slug}/application/submit`, { method: 'POST', body: JSON.stringify({ revision: revision.current, program_revision: program.revision }) }); setDirty(false); router.push(`/programs/${program.slug}/application`); router.refresh(); }
  catch (cause) { setNeedsSignIn(cause instanceof ApiClientError && cause.code === 'UNAUTHORIZED'); setError(cause instanceof Error ? cause.message : 'Your application could not be submitted. Please try again.'); }
  finally { setBusy(false); }
 }
 if (!data.accepting_applications) return <div className="program-outcome"><h2>Applications are currently closed.</h2><p>Your saved application is still available. New submissions and draft changes are paused.</p><Link className="program-button" href={`/programs/${program.slug}/application`}>View your application</Link></div>;
 return <div className="program-application-layout"><aside className="program-step-guide"><nav aria-label="Application steps"><ol role="list">{steps.map((item, position) => <li key={item.id} aria-current={item.id === step ? 'step' : undefined}><span aria-hidden="true">{position + 1}</span>{position < index ? <button type="button" className="program-step-link" disabled={busy} onClick={() => void navigate(`/programs/${program.slug}/apply/${item.id}`)}>{item.label}</button> : item.label}</li>)}</ol></nav><div className="program-side-note"><h2>Make it yours.</h2><p>Tell us about the people you want to bring together and the event you have in mind.</p><p>You can save your draft before submitting. Decisions appear in your application after results are released.</p></div></aside><form className="program-form" onSubmit={event => void submit(event)}>
  <div className="program-form-heading"><span>Step {index + 1} of 3</span><h2>{step === 'profile' ? 'First, a little about you.' : step === 'proposal' ? 'What would you bring to life?' : 'Review your application.'}</h2><p>{step === 'profile' ? 'Introduce yourself and share the best details for reaching you.' : step === 'proposal' ? 'An idea, a community and a reason to meet. Start there.' : 'Check your details before sending your application to the program team.'}</p></div>
  <fieldset className="program-input-fieldset" disabled={busy} aria-label="Application answers">
  {step === 'profile' && <ProgramProfileFields values={values} email={user.email} minimumAge={program.minimum_age} ambassador={program.slug === 'ambassadors'} busy={busy} onChange={change} />}
  {step === 'proposal' && <div className="program-fields"><label htmlFor="program-proposal">Your event idea<textarea id="program-proposal" name="proposal" rows={8} required minLength={30} maxLength={5000} value={values.proposal} onChange={e => change('proposal', e.target.value)} /><span>Who would it bring together? What would happen? Why does it matter to you? At least 30 characters.</span></label>{program.questions.map(q => <div key={q.id}>{q.kind === 'choice' ? <label htmlFor={`program-choice-${q.id}`}>{q.label}{q.required ? ' *' : ' (optional)'}<Select id={`program-choice-${q.id}`} disabled={busy} label={`${q.label}${q.required ? ' *' : ''}`} name={q.id} value={values.answers[q.id] || ''} options={(q.options || []).map(value => ({ value, label: value }))} placeholder="Choose an answer" required={q.required} onValueChange={value => change('answers', { ...values.answers, [q.id]: value })} /></label> : <label htmlFor={`program-question-${q.id}`}>{q.label}{q.required ? ' *' : ' (optional)'}{q.kind === 'long_text' ? <textarea id={`program-question-${q.id}`} name={q.id} rows={5} required={q.required} maxLength={q.max_length} value={values.answers[q.id] || ''} onChange={e => change('answers', { ...values.answers, [q.id]: e.target.value })} /> : <input id={`program-question-${q.id}`} name={q.id} type={q.kind === 'url' ? 'url' : 'text'} required={q.required} maxLength={q.max_length} value={values.answers[q.id] || ''} onChange={e => change('answers', { ...values.answers, [q.id]: e.target.value })} />}</label>}{(q.help || q.domain) && <p className="program-field-help">{q.help}{q.domain && ` Use ${q.domain} or one of its subdomains.`}</p>}</div>)}</div>}
  {step === 'review' && <><ProgramReviewAnswers program={program} values={values} email={user.email} /><div className="program-review-commitment"><h3>Your role in the program</h3><p className="program-prose">{program.responsibilities}</p><label className="program-check"><Checkbox disabled={busy} required checked={confirmed} onCheckedChange={setConfirmed} />I have checked my application and understand the program responsibilities.</label><p>We will email you a confirmation after you apply and share updates by email.</p></div></>}
  </fieldset>
  {error && <p className="program-error" role="alert">{error}</p>}{needsSignIn && <p className="program-auth-recovery"><a href={`/login?next=${encodeURIComponent(`/programs/${program.slug}/application`)}`} target="_blank" rel="noopener noreferrer">Sign in in another tab</a>, then return here and retry. Your answers are still on this page.</p>}{notice && <p className="program-success" role="status">{notice}</p>}
  <div className="program-form-actions"><button className="program-text-link" type="button" disabled={busy} onClick={() => void navigate(index > 0 ? `/programs/${program.slug}/apply/${steps[index - 1].id}` : `/programs/${program.slug}`)}>{index > 0 ? 'Back' : 'Program details'}</button><div>{step !== 'review' && <button className="program-button program-button-secondary" type="button" disabled={busy} onClick={() => void save()}>Save draft</button>}<button className="program-button" type="submit" disabled={busy || step === 'review' && !confirmed}>{busy ? 'Saving…' : step === 'review' ? 'Submit application' : 'Save & continue'}</button></div></div>
 </form></div>;
}
