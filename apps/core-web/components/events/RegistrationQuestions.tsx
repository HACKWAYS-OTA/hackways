import { Checkbox } from '@hackways/ui/checkbox';
import { RadioGroup } from '@hackways/ui/radio-group';
import type { RegistrationQuestion, RegistrationAnswer } from '@/types/api';
import { Textarea } from '@/components/ui/Textarea';
export function RegistrationQuestions({ questions, answers, onChange, disabled }: { questions: RegistrationQuestion[]; answers: Record<string, RegistrationAnswer>; onChange: (answer: RegistrationAnswer) => void; disabled: boolean }) {
  return <div className="registration-custom-questions">{questions.map(q => {
    const answer = answers[q.public_code] || { code: q.public_code, text: '', checked: false };
    const id = `answer-${q.public_code}`, helpId = `${id}-help`;
    const label = <>{q.label}{!q.required && <span> (optional)</span>}</>;
    if (q.kind === 'select') return <fieldset className="registration-answer-choice" key={q.public_code}><legend>{label}</legend>{q.description && <p id={helpId} className="field-help">{q.description}</p>}<RadioGroup label={q.label} name={id} options={q.options.map(option => ({ value: option, label: option }))} value={answer.text} required={q.required} disabled={disabled} describedBy={q.description ? helpId : undefined} layout="stack" onValueChange={text => onChange({ ...answer, text })} />{!q.required && answer.text && <button type="button" className="quiet-link" disabled={disabled} onClick={() => onChange({ ...answer, text: '' })}>Clear selection</button>}</fieldset>;
    if (q.kind === 'checkbox') return <div className="field" key={q.public_code}><label className="registration-answer-check"><Checkbox id={id} checked={answer.checked} required={q.required} disabled={disabled} aria-describedby={q.description ? helpId : undefined} onCheckedChange={checked => onChange({ ...answer, checked })} /><span>{label}</span></label>{q.description && <p id={helpId} className="field-help">{q.description}</p>}</div>;
    return <div className="field" key={q.public_code}><label htmlFor={id}>{label}</label>{q.kind === 'textarea' ? <Textarea id={id} value={answer.text} maxLength={2000} required={q.required} disabled={disabled} aria-describedby={q.description ? helpId : undefined} onChange={e => onChange({ ...answer, text: e.target.value })} /> : <input id={id} className="app-input" type={q.kind === 'url' ? 'url' : q.kind === 'phone' ? 'tel' : 'text'} inputMode={q.kind === 'number' ? 'decimal' : undefined} pattern={q.kind === 'url' ? 'https://.*' : undefined} value={answer.text} maxLength={2000} required={q.required} disabled={disabled} aria-describedby={q.description ? helpId : undefined} onChange={e => onChange({ ...answer, text: e.target.value })} />}{q.description && <p id={helpId} className="field-help">{q.description}</p>}</div>;
  })}</div>;
}
