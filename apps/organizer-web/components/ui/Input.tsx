import { useId, type InputHTMLAttributes } from 'react';
export interface InputProps extends InputHTMLAttributes<HTMLInputElement> { label?: string; error?: string; helperText?: string }
export function Input({ label, error, helperText, id, className = '', ...props }: InputProps) {
  const generatedId = useId();
  const inputId = id || generatedId;
  const message = error || helperText;
  return <div className="field">{label && <label htmlFor={inputId}>{label}</label>}<input {...props} id={inputId} className={`app-input ${className}`} aria-invalid={error ? true : undefined} aria-describedby={message ? `${inputId}-message` : props['aria-describedby']} />{message && <p id={`${inputId}-message`} className={error ? 'field-error' : 'field-help'}>{message}</p>}</div>;
}
