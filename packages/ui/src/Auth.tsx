'use client';

import { useId, useState, type ButtonHTMLAttributes, type InputHTMLAttributes, type ReactNode } from 'react';
import { HugeiconsIcon } from '@hugeicons/react';
import GoogleIcon from '@hugeicons/core-free-icons/GoogleIcon';
import ArrowRight01Icon from '@hugeicons/core-free-icons/ArrowRight01Icon';
import ViewIcon from '@hugeicons/core-free-icons/ViewIcon';
import ViewOffSlashIcon from '@hugeicons/core-free-icons/ViewOffSlashIcon';
import { Toggle } from './Toggle';

export function AuthLayout({ title, description, homeHref, children }: { title: string; description?: string; homeHref: string; children: ReactNode }) {
  return <section className="hw-auth-view">
    <div className="hw-auth-brand">
      <a className="hw-auth-wordmark" href={homeHref} aria-label="Hackways home">hackways.</a>
      <p className="hw-auth-statement">Good things happen<br />when people meet.</p>
    </div>
    <div className="hw-auth-form-side"><div className="hw-auth-content">
      <div className="hw-auth-intro"><h1>{title}</h1>{description && <p>{description}</p>}</div>
      {children}
    </div></div>
  </section>;
}

export function AuthGoogleButton(props: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button {...props} type="button" className={`hw-auth-google ${props.className || ''}`}>
    <HugeiconsIcon icon={GoogleIcon} size={21} strokeWidth={1.8} aria-hidden="true" />Continue with Google
  </button>;
}

export function AuthDivider() {
  return <div className="hw-auth-divider"><span>or use email</span></div>;
}

export function AuthSubmit({ children, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button {...props} type="submit" className={`hw-auth-submit ${props.className || ''}`}>
    <span>{children}</span><HugeiconsIcon icon={ArrowRight01Icon} size={19} strokeWidth={1.8} aria-hidden="true" />
  </button>;
}

export function AuthPasswordInput({ label = 'Password', helperText, id, ...props }: Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> & { label?: string; helperText?: string }) {
  const generatedId = useId();
  const inputId = id || generatedId;
  const [visible, setVisible] = useState(false);
  return <div className="field hw-auth-password">
    <label htmlFor={inputId}>{label}</label>
    <div className="hw-auth-password-control">
      <input {...props} id={inputId} type={visible ? 'text' : 'password'} className="app-input" aria-describedby={helperText ? `${inputId}-help` : props['aria-describedby']} />
      <Toggle className="hw-auth-reveal" type="button" pressed={visible} onPressedChange={setVisible}
        aria-label={visible ? 'Hide password' : 'Show password'} aria-controls={inputId} disabled={props.disabled}>
        <HugeiconsIcon icon={visible ? ViewOffSlashIcon : ViewIcon} size={20} strokeWidth={1.8} aria-hidden="true" />
      </Toggle>
    </div>
    {helperText && <p className="field-help" id={`${inputId}-help`}>{helperText}</p>}
  </div>;
}
