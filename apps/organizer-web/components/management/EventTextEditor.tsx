'use client';

import { useEffect, useId, useLayoutEffect, useRef, useState } from 'react';

export function EventTextEditor({ label, name, value, onChange, maxLength, required = false, disabled = false, hideLabel = false }: { label: string; name: string; value: string; onChange: (value: string) => void; maxLength: number; required?: boolean; disabled?: boolean; hideLabel?: boolean }) {
  const id = useId(), input = useRef<HTMLTextAreaElement>(null);
  const [invalid, setInvalid] = useState(false);
  function resize() { if (input.current?.offsetParent) { input.current.style.height = '0px'; input.current.style.height = `${Math.max(88, input.current.scrollHeight)}px`; } }
  useLayoutEffect(resize, [value]);
  useEffect(() => {
    const disclosure = input.current?.closest('details');
    disclosure?.addEventListener('toggle', resize); window.addEventListener('resize', resize);
    return () => { disclosure?.removeEventListener('toggle', resize); window.removeEventListener('resize', resize); };
  }, []);
  return <div className="event-text-editor" data-invalid={invalid || undefined}>
    <label className={hideLabel ? 'visually-hidden' : 'event-text-label'} htmlFor={id}>{label}</label>
    <textarea id={id} ref={input} name={name} rows={3} value={value} onChange={event => { onChange(event.target.value); setInvalid(false); }} maxLength={maxLength} required={required} disabled={disabled} aria-invalid={invalid || undefined} aria-describedby={invalid ? `${id}-error` : undefined} onInvalid={event => { event.preventDefault(); setInvalid(true); requestAnimationFrame(() => input.current?.focus()); }} />
    {invalid && <p id={`${id}-error`} className="event-control-error" role="alert">Add {label.toLowerCase()} before saving.</p>}
  </div>;
}
