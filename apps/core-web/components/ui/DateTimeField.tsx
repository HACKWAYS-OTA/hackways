'use client';

import { useId, useRef, useState, type KeyboardEvent } from 'react';
import './date-time-field.css';

const pad = (n: number) => String(n).padStart(2, '0');
const dateKey = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
const readDate = (s: string) => new Date(`${s}T12:00:00`);

function TimeWheel({ id, label, count, value, onChange, period = false }: { id: string; label: string; count: number; value: number; onChange: (value: number) => void; period?: boolean }) {
  const ref = useRef<HTMLDivElement>(null);
  function key(e: KeyboardEvent<HTMLDivElement>) {
    const next = e.key === 'ArrowDown' ? (value + 1) % count : e.key === 'ArrowUp' ? (value + count - 1) % count : e.key === 'Home' ? 0 : e.key === 'End' ? count - 1 : null;
    if (next === null) return;
    e.preventDefault(); onChange(next);
    ref.current?.querySelector<HTMLElement>(`[data-index="${next}"]`)?.scrollIntoView({ block: 'nearest' });
  }
  return <div ref={ref} className={`time-wheel${period ? ' time-wheel-period' : ''}`} role="listbox" tabIndex={0} aria-label={label} aria-activedescendant={`${id}-${value}`} onKeyDown={key} onScroll={e => {
    if (period) return;
    const wheel = e.currentTarget, cycle = (wheel.scrollHeight - 8) / 3;
    if (wheel.scrollTop < cycle / 2) wheel.scrollTop += cycle;
    else if (wheel.scrollTop > cycle * 1.5) wheel.scrollTop -= cycle;
  }}>{Array.from({ length: period ? count : count * 3 }, (_, index) => {
    const n = index % count, primary = period || Math.floor(index / count) === 1;
    return <div role="option" key={index} id={primary ? `${id}-${n}` : undefined} data-index={primary ? n : undefined} data-highlight={value === n || undefined} aria-selected={primary && value === n} onClick={() => onChange(n)}>{period ? ['AM', 'PM'][n] : label === 'Hour' ? pad(n + 1) : pad(n)}</div>;
  })}</div>;
}

/** Local wall-clock value. The submitting form owns time-zone conversion. */
export function DateTimeField({ name, label, defaultValue = '', required = false, disabled = false }: {
  name: string; label: string; defaultValue?: string; required?: boolean; disabled?: boolean;
}) {
  const id = useId();
  const dialog = useRef<HTMLDialogElement>(null);
  const dateTrigger = useRef<HTMLButtonElement>(null);
  const timeTrigger = useRef<HTMLButtonElement>(null);
  const [date, setDate] = useState(defaultValue.slice(0, 10));
  const [time, setTime] = useState(defaultValue.slice(11));
  const [panel, setPanel] = useState<'date' | 'time'>('date');
  const [month, setMonth] = useState(() => defaultValue ? readDate(defaultValue.slice(0, 10)) : new Date());
  const [hour, setHour] = useState(Number(defaultValue.slice(11, 13) || 9));
  const [minute, setMinute] = useState(Number(defaultValue.slice(14, 16) || 0));
  const [invalid, setInvalid] = useState(false);
  const complete = !!date && !!time;
  const value = complete ? `${date}T${time}` : '';
  const firstDay = new Date(month.getFullYear(), month.getMonth(), 1);
  const days = Array.from({ length: 42 }, (_, i) => new Date(month.getFullYear(), month.getMonth(), 1 - firstDay.getDay() + i));

  function open(next: 'date' | 'time') {
    setPanel(next);
    if (next === 'date') setMonth(date ? readDate(date) : new Date());
    else { setHour(Number(time.slice(0, 2) || 9)); setMinute(Number(time.slice(3, 5) || 0)); }
    dialog.current?.showModal();
    requestAnimationFrame(() => dialog.current?.querySelectorAll<HTMLElement>('.time-wheel').forEach(wheel => {
      const selected = wheel.querySelector<HTMLElement>('[aria-selected=true]');
      if (selected && !wheel.classList.contains('time-wheel-period')) wheel.scrollTop = selected.offsetTop - wheel.offsetTop - 88;
    }));
  }
  function close() {
    dialog.current?.close();
    (panel === 'date' ? dateTrigger : timeTrigger).current?.focus();
  }
  function choose(d: Date) { setDate(dateKey(d)); setInvalid(false); close(); }
  function moveDay(e: KeyboardEvent<HTMLButtonElement>, day: Date) {
    const offsets: Record<string, number> = { ArrowLeft: -1, ArrowRight: 1, ArrowUp: -7, ArrowDown: 7, Home: -day.getDay(), End: 6 - day.getDay() };
    let next: Date;
    if (e.key === 'PageUp' || e.key === 'PageDown') {
      const targetMonth = day.getMonth() + (e.key === 'PageUp' ? -1 : 1);
      const last = new Date(day.getFullYear(), targetMonth + 1, 0).getDate();
      next = new Date(day.getFullYear(), targetMonth, Math.min(day.getDate(), last));
    } else if (e.key in offsets) next = new Date(day.getFullYear(), day.getMonth(), day.getDate() + offsets[e.key]);
    else return;
    e.preventDefault(); setMonth(next);
    requestAnimationFrame(() => dialog.current?.querySelector<HTMLButtonElement>(`[data-day="${dateKey(next)}"]`)?.focus());
  }

  return <div className="date-time-field" role="group" aria-labelledby={`${id}-label`} data-invalid={invalid || undefined}>
    <span id={`${id}-label`} className="visually-hidden">{label}</span>
    <input type="hidden" name={name} value={value} disabled={disabled} />
    {(required || date || time) && <input className="picker-validation" tabIndex={-1} aria-label={`${label}, date and time required`} value={value} onChange={() => undefined} required disabled={disabled} onInvalid={e => { e.preventDefault(); setInvalid(true); if (!document.querySelector('dialog[open]')) open(date ? 'time' : 'date'); }} />}
    <button type="button" ref={dateTrigger} disabled={disabled} className="date-block" aria-haspopup="dialog" aria-label={`${label} date${date ? `: ${date}` : ''}`} onClick={() => open('date')}>{date ? readDate(date).toLocaleDateString('en', { weekday: 'short', day: 'numeric', month: 'short' }) : 'Choose date'}</button>
    <button type="button" ref={timeTrigger} disabled={disabled} className="time-block" aria-haspopup="dialog" aria-label={`${label} time${time ? `: ${time.slice(0, 5)}` : ''}`} onClick={() => open('time')}>{time ? `${pad(Number(time.slice(0, 2)) % 12 || 12)}:${time.slice(3, 5)} ${Number(time.slice(0, 2)) >= 12 ? 'PM' : 'AM'}` : 'Set time'}</button>
    <dialog ref={dialog} className={`date-time-dialog${panel === 'time' ? ' time-picker-dialog' : ''}`} aria-labelledby={`${id}-title`} onCancel={() => { (panel === 'date' ? dateTrigger : timeTrigger).current?.focus(); }} onClick={e => { if (e.target === dialog.current) close(); }}>
      <div className="picker-body"><header><h2 id={`${id}-title`}>{label} {panel === 'date' ? 'date' : 'time'}</h2><button type="button" onClick={close}>Close</button></header>
        {panel === 'date' ? <>
          <div className="calendar-heading"><button type="button" aria-label="Previous month" onClick={() => setMonth(new Date(month.getFullYear(), month.getMonth() - 1, 1))}>Previous</button><h3 aria-live="polite">{month.toLocaleDateString('en', { month: 'long', year: 'numeric' })}</h3><button type="button" aria-label="Next month" onClick={() => setMonth(new Date(month.getFullYear(), month.getMonth() + 1, 1))}>Next</button></div>
          <div className="calendar-weekdays" aria-hidden="true">{['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'].map(day => <span key={day}>{day}</span>)}</div>
          <div className="calendar-days" role="group" aria-label="Choose a day">{days.map(day => <button type="button" key={dateKey(day)} data-day={dateKey(day)} className={day.getMonth() !== month.getMonth() ? 'outside-month' : ''} aria-pressed={date === dateKey(day)} aria-current={dateKey(day) === dateKey(new Date()) ? 'date' : undefined} aria-label={day.toLocaleDateString('en', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })} onKeyDown={e => moveDay(e, day)} onClick={() => choose(day)}>{day.getDate()}</button>)}</div>
          <p className="picker-hint">Arrow keys move between days. Page Up and Page Down change months.</p>
        </> : <>
          <p className="time-preview" aria-live="polite">{pad(hour % 12 || 12)}:{pad(minute)} <span>{hour >= 12 ? 'PM' : 'AM'}</span></p>
          <div className="time-wheel-labels" aria-hidden="true"><span>Hour</span><span>Minute</span><span>Period</span></div>
          <div className="time-wheels">
            <TimeWheel id={`${id}-hour`} label="Hour" count={12} value={(hour % 12 || 12) - 1} onChange={n => setHour((n + 1) % 12 + Math.floor(hour / 12) * 12)} />
            <TimeWheel id={`${id}-minute`} label="Minute" count={60} value={minute} onChange={setMinute} />
            <TimeWheel id={`${id}-period`} label="AM or PM" count={2} value={Math.floor(hour / 12)} onChange={n => setHour(hour % 12 + n * 12)} period />
          </div>
          <button type="button" className="picker-done" onClick={() => { setTime(`${pad(hour)}:${pad(minute)}`); setInvalid(false); close(); }}>Set time</button>
        </>}
        <footer>{!required && <button type="button" onClick={() => { setDate(''); setTime(''); close(); }}>Clear date and time</button>}{invalid && <p role="alert">Choose both a date and a time.</p>}</footer>
      </div>
    </dialog>
  </div>;
}
