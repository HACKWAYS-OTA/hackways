"use client";
import { useEffect, useRef, useState, type FormEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import type { Event, Registration, User, RegistrationFormConfig, RegistrationAnswer } from '@/types/api';
import { Input } from '@/components/ui/Input';
import { Textarea } from '@/components/ui/Textarea';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Icon } from '@/components/ui/Icon';
import { fetchApi, ApiClientError } from '@/lib/api';
import { eventDate, eventPath, eventZone, registrationState } from '@/lib/events';

import { RegistrationQuestions } from './RegistrationQuestions';

export function RegistrationForm({ event, user, initialForm }: { event: Event; user: User | null; initialForm: RegistrationFormConfig }) {
  const [form, setForm] = useState(initialForm);
  const [answers, setAnswers] = useState<Record<string, RegistrationAnswer>>({});
  const [needsReload, setNeedsReload] = useState(false);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [confirmed, setConfirmed] = useState<Registration | null>(null);
  const submitting = useRef(false);
  const attempt = useRef<{ payload: string; key: string } | null>(null);
  const feedback = useRef<HTMLDivElement>(null);
  const router = useRouter();
  const state = registrationState(event);
  const awaitingApproval = confirmed?.status === 'pending_approval';
  const waitlisted = confirmed?.status === 'waitlisted';
  const joiningWaitlist = state.label === 'Join waiting list';
  const heading = <header className="event-detail-heading"><h1>{confirmed ? waitlisted ? 'Youâ€™re on the waiting list.' : awaitingApproval ? 'Your application is submitted.' : 'Your place is reserved.' : joiningWaitlist ? 'Join the waiting list.' : event.registration_mode === 'approval' ? 'Request a place.' : 'Reserve your place.'}</h1><p className="registration-event-title">{event.title}</p><p className="registration-event-date">{eventDate(event, { month: 'long', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })} {eventZone(event)}</p></header>;
  useEffect(() => { if (error || confirmed) feedback.current?.focus({ preventScroll: true }); }, [error, confirmed]);

  async function reloadQuestions() {
    const result = await fetchApi<{ data: RegistrationFormConfig }>(`/events/by-slug/${encodeURIComponent(event.slug)}/registration-form`, { cache: 'no-store' });
    setForm(result.data); setNeedsReload(false);
  }

  async function submit(action: FormEvent<HTMLFormElement>) {
    action.preventDefault();
    if (submitting.current || state.disabled || needsReload) return;
    if (!user && !email.trim()) { setError('Enter your email address to reserve a place.'); return; }
    submitting.current = true; setLoading(true); setError(null);
    try {
      const payload = JSON.stringify({ event_id: event.id, guest_name: user?.display_name || name.trim(), guest_email: user?.email || email.trim(), notes: notes.trim(), form_revision: form.revision, answers: form.questions.map(q => answers[q.public_code] || { code: q.public_code, text: '', checked: false }), ...(joiningWaitlist ? { waitlist: true } : {}) });
      if (attempt.current?.payload !== payload) attempt.current = { payload, key: crypto.randomUUID() };
      const response = await fetchApi<{ data: Registration }>('/rsvps', { method: 'POST', headers: { 'Idempotency-Key': attempt.current.key }, body: payload });
      if (!response.data?.id || response.data.event_id !== event.id || !['confirmed', 'pending_approval', 'waitlisted'].includes(response.data.status)) throw new Error('The registration response could not be verified. Check its current status before retrying.');
      setConfirmed(response.data);
      router.refresh();
    } catch (failure) {
      setError(failure instanceof Error ? failure.message : 'Your reservation could not be confirmed. Please try again.');
      if (failure instanceof ApiClientError && failure.code === 'FORM_CHANGED') {
        setNeedsReload(true);
        try { await reloadQuestions(); setError('The organizer updated these questions. Your answers are kept. Review the form and submit again.'); }
        catch { setError('The questions changed, but the latest form could not be loaded. Reload the questions to continue. Your entries are still here.'); }
      }
    }
    finally { submitting.current = false; setLoading(false); }
  }

  if (confirmed) return <>{heading}<div className="registration-success registration-page-success" ref={feedback} tabIndex={-1} role="status"><span className="success-icon"><Icon name="check" size={32} /></span><Badge variant={awaitingApproval || waitlisted ? 'warning' : 'success'}>{waitlisted ? 'On the waiting list' : awaitingApproval ? 'Awaiting organizer approval' : 'RSVP confirmed'}</Badge>{waitlisted && <p>No place is reserved yet. Check My events for an offer when a place becomes available.</p>}{awaitingApproval && <p>The organizer will review your application. A place has not been reserved yet.</p>}{(confirmed.guest_name || confirmed.guest_email) && <p className="confirmed-attendee">{awaitingApproval || waitlisted ? 'Submitted for' : 'Reserved for'} {confirmed.guest_name || confirmed.guest_email}</p>}<div className="reservation-reference"><span>Registration reference</span><span>{confirmed.public_code}</span></div><Link className="app-button app-button-primary" href={eventPath(event)}>Back to event <Icon name="arrow" size={17} /></Link><Link className="detail-text-link" href="/my-events">View my events</Link></div></>;
  if (state.disabled) return <>{heading}<section className="event-registration-panel"><h2>{state.label}</h2><p>Registration is not available at the moment.</p><Link className="detail-text-link" href={eventPath(event)}>Back to event details</Link></section></>;
  if (!user && (event.allow_guest === false || event.registration_mode === "approval" || joiningWaitlist)) return <>{heading}<section className="event-registration-panel"><h2>Sign in to join.</h2><p>This event requires an account so you can keep track of your registration.</p><Link className="app-button app-button-primary" href={`/login?next=${encodeURIComponent(eventPath(event) + "/register")}`}>Sign in or create an account</Link></section></>;
  return <>{heading}<form onSubmit={submit} className="registration-form event-page-form"><p className="registration-intro">{user ? `Registering as ${user.display_name} (${user.email}).` : <>Share your details with the event organizer. <Link className="quiet-link" href={`/login?next=${encodeURIComponent(eventPath(event) + "/register")}`}>Sign in to keep this in My events</Link></>}</p>{error && <div className="form-alert" role="alert" ref={feedback} tabIndex={-1}><Icon name="alert" size={19} /><p>{error}</p></div>}
    {needsReload && <button className="app-button app-button-outline" type="button" disabled={loading} onClick={async () => { setLoading(true); try { await reloadQuestions(); setError('Questions updated. Review your answers before submitting again.'); } catch (err) { setError(err instanceof Error ? err.message : 'The latest questions could not be loaded.'); } finally { setLoading(false); } }}>Reload questions</button>}
    {!user && <><Input label={form.name_required ? 'Your name' : 'Your name (optional)'} required={form.name_required} name="name" autoComplete="name" value={name} maxLength={200} onChange={(action) => setName(action.target.value)} disabled={loading} />
    <Input label="Email address" name="email" type="email" autoComplete="email" inputMode="email" required maxLength={320} value={email} onChange={(action) => setEmail(action.target.value)} disabled={loading} helperText="Use the same email for any reservation retries." /></>}
    {user && form.name_required && !user.display_name.trim() && <p className="form-alert">This event requires your name. <Link href="/account">Add your name to your account</Link> before registering.</p>}
    <RegistrationQuestions questions={form.questions} answers={answers} onChange={answer => setAnswers(current => ({ ...current, [answer.code]: answer }))} disabled={loading} />
    <div className="field"><label htmlFor="registration-notes">A note for the organizer <span>(optional)</span></label><Textarea id="registration-notes" name="notes" maxLength={2000} value={notes} onChange={(action) => setNotes(action.target.value)} disabled={loading} aria-describedby="registration-notes-help" /><p id="registration-notes-help" className="field-help">Anything you would like them to know before the event.</p></div>
    <div className="registration-actions"><Button variant="outline" disabled={loading} onClick={() => router.push(eventPath(event))}>Back</Button><Button type="submit" isLoading={loading} disabled={needsReload}>{loading ? 'Submitting...' : joiningWaitlist ? 'Join waiting list' : event.registration_mode === 'approval' ? 'Submit application' : 'Confirm RSVP'}{!loading && <Icon name="arrow" size={17} />}</Button></div><p className="registration-note">Availability is checked when you confirm your RSVP.</p>
  </form></>;
}
