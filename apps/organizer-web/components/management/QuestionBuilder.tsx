'use client';
import { Checkbox } from '@hackways/ui/checkbox';
import { RadioGroup } from '@hackways/ui/radio-group';
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import type { RegistrationFormConfig, RegistrationQuestion, QuestionKind } from '@/types/api';
import { fetchApi, ApiClientError } from '@/lib/api';
import { Icon } from '@/components/ui/Icon';

const kinds: [QuestionKind, string][] = [['text', 'Short answer'], ['textarea', 'Long answer'], ['select', 'Choose one'], ['checkbox', 'Checkbox'], ['number', 'Number'], ['url', 'Website'], ['phone', 'Phone number']];
type Draft = RegistrationQuestion & { localKey: string };
const drafts = (questions: RegistrationQuestion[]) => questions.map((q, i) => ({ ...q, localKey: q.public_code || `new-${i}` }));
export function QuestionBuilder({ eventId, initial, editable }: { eventId: string; initial: RegistrationFormConfig; editable: boolean }) {
  const router = useRouter(), sequence = useRef(initial.questions.length), lock = useRef(false);
  const [questions, setQuestions] = useState<Draft[]>(() => drafts(initial.questions));
  const [nameRequired, setNameRequired] = useState(initial.name_required), [revision, setRevision] = useState(initial.revision);
  const [baseline, setBaseline] = useState(JSON.stringify({ name: initial.name_required, questions: drafts(initial.questions) }));
  const [expanded, setExpanded] = useState<string | null>(null), [remove, setRemove] = useState<string | null>(null);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [saved, setSaved] = useState(false), [conflict, setConflict] = useState(false);
  const dirty = JSON.stringify({ name: nameRequired, questions }) !== baseline;
  function update(key: string, patch: Partial<RegistrationQuestion>) { setSaved(false); setQuestions(current => current.map(q => q.localKey === key ? { ...q, ...patch } : q)); }
  function move(index: number, by: number) { setSaved(false); setQuestions(current => { const result = [...current]; [result[index], result[index + by]] = [result[index + by], result[index]]; return result; }); }
  function add() { const localKey = `new-${++sequence.current}`; setQuestions(current => [...current, { localKey, public_code: '', label: '', description: '', kind: 'text', required: false, options: [] }]); setExpanded(localKey); setSaved(false); }
  async function save(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); if (lock.current || !editable) return; lock.current = true; setBusy(true); setError(''); setSaved(false); setConflict(false);
    try {
      const result = await fetchApi<{ data: RegistrationFormConfig }>(`/events/${eventId}/registration-form`, { method: 'PUT', body: JSON.stringify({ revision, name_required: nameRequired, questions: questions.map(({ localKey: _key, ...q }) => q) }) });
      const next = drafts(result.data.questions); setQuestions(next); setRevision(result.data.revision); setNameRequired(result.data.name_required); setBaseline(JSON.stringify({ name: result.data.name_required, questions: next })); setExpanded(null); setSaved(true); router.refresh();
    } catch (err) { setError(err instanceof Error ? err.message : 'Your questions could not be saved. Your edits are still here.'); setConflict(err instanceof ApiClientError && err.code === 'FORM_CHANGED'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <form className="host-question-builder" onSubmit={save} aria-busy={busy}>
    <div className="host-section-heading"><div><h2>Ask only what matters.</h2><p className="host-muted">Questions guests answer when they register. Past answers stay as submitted.</p></div><span className="host-question-limit">{questions.length} / 20 questions</span></div>
    {!editable && <p className="journey-notice">Questions are read-only after an event starts or is cancelled.</p>}
    <fieldset disabled={busy} className="host-form-fields"><legend className="visually-hidden">Configure guest questions</legend>
      <section className="host-identity-fields" aria-labelledby="identity-heading"><h3 id="identity-heading">Their details</h3><div><span><strong>Email address</strong><small>Always required for registration</small></span><span className="host-required-label">Required</span></div><label><span><strong>Name</strong><small>Signed-in guests use their account name</small></span><span className="host-toggle-label"><Checkbox disabled={busy || !editable} checked={nameRequired} onCheckedChange={checked => { setNameRequired(checked); setSaved(false); }} />Required</span></label></section>
      <section aria-labelledby="question-heading" className="host-question-section"><h3 id="question-heading">Your questions</h3>
        {questions.length === 0 && <div className="host-question-empty"><Icon name="people" size={25} /><p>No extra questions yet.</p><small>Keep registration short, or ask for the details you need to welcome your guests.</small></div>}
        <ol className="host-question-list">{questions.map((q, index) => <li key={q.localKey} className={expanded === q.localKey ? 'is-editing' : ''}>
          <div className="host-question-row"><button type="button" className="host-question-open" aria-expanded={expanded === q.localKey} aria-controls={`question-${q.localKey}`} onClick={() => setExpanded(expanded === q.localKey ? null : q.localKey)}><span>{q.label || 'Untitled question'}<small>{kinds.find(([key]) => key === q.kind)?.[1]} · {q.required ? 'Required' : 'Optional'}</small></span><Icon name="down" size={16} /></button>{editable && <div className="host-question-order"><button type="button" disabled={index === 0} aria-label={`Move question ${index + 1} up`} onClick={() => move(index, -1)}><Icon name="down" className="host-arrow-up" size={16} /></button><button type="button" disabled={index === questions.length - 1} aria-label={`Move question ${index + 1} down`} onClick={() => move(index, 1)}><Icon name="down" size={16} /></button></div>}</div>
          <fieldset disabled={!editable} className="host-question-editor" id={`question-${q.localKey}`} hidden={expanded !== q.localKey}><legend className="visually-hidden">Question details</legend>
            <div className="field"><label htmlFor={`label-${q.localKey}`}>Question</label><input id={`label-${q.localKey}`} className="app-input" value={q.label} maxLength={180} onChange={e => update(q.localKey, { label: e.target.value })} /></div>
            <fieldset className="hw-choice-field"><legend>Answer type</legend><RadioGroup label="Answer type" name={`kind-${q.localKey}`} value={q.kind} disabled={busy || !editable} options={kinds.map(([value, label]) => ({ value, label }))} onValueChange={kind => update(q.localKey, { kind, options: kind === 'select' ? q.options.length ? q.options : ['', ''] : [] })} /></fieldset>
            <div className="field"><label htmlFor={`help-${q.localKey}`}>Help text <span>(optional)</span></label><input id={`help-${q.localKey}`} className="app-input" value={q.description} maxLength={500} onChange={e => update(q.localKey, { description: e.target.value })} /></div>
            {q.kind === 'select' && <div className="field"><label htmlFor={`options-${q.localKey}`}>Choices — one per line</label><textarea id={`options-${q.localKey}`} className="app-input" rows={4} maxLength={2020} value={q.options.join('\n')} onChange={e => update(q.localKey, { options: e.target.value.split('\n') })} /><p className="field-help">2–20 different choices, up to 100 characters each.</p></div>}
            <div className="host-question-editor-footer"><label className="check-label"><Checkbox checked={q.required} disabled={busy || !editable} onCheckedChange={required => update(q.localKey, { required })} />Answer is required</label><button type="button" className="quiet-link" onClick={() => setRemove(q.localKey)}>Remove question</button></div>
            {remove === q.localKey && <div className="host-question-remove"><p>Remove this question from new registrations? Past answers are kept.</p><button type="button" className="quiet-link" onClick={() => { setQuestions(current => current.filter(item => item.localKey !== q.localKey)); setRemove(null); setSaved(false); }}>Remove question</button><button type="button" className="quiet-link" onClick={() => setRemove(null)}>Keep it</button></div>}
          </fieldset></li>)}</ol>
        {editable && <button type="button" className="app-button app-button-outline" disabled={questions.length >= 20} onClick={add}><Icon name="add" size={17} />Add a question</button>}
      </section>
    </fieldset>
    {error && <div className="journey-error" role="alert"><p>{error}</p>{conflict && <button type="button" className="quiet-link" onClick={() => router.refresh()}>Load latest questions (replaces your edits)</button>}</div>}
    {saved && <p className="journey-notice" role="status">Questions saved. New registrations will use this form.</p>}
    {editable && <footer className="host-save"><span>{dirty ? 'You have unsaved changes.' : 'Your registration form is up to date.'}</span><button className="app-button app-button-primary" disabled={busy || !dirty}>{busy ? 'Saving…' : 'Save questions'}</button></footer>}
  </form>;
}
