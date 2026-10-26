'use client';
/* Image previews use browser object URLs in this framework-independent control;
   Next's image proxy cannot read them, and uploads go directly to Azure. */
/* eslint-disable @next/next/no-img-element */
import { useEffect, useId, useRef, useState } from 'react';
import { HugeiconsIcon } from '@hugeicons/react';
import Image01Icon from '@hugeicons/core-free-icons/Image01Icon';
import './media-field.css';
import { ImageCropDialog } from './ImageCropDialog';

export function MediaField({ label, file, initialUrl, removed, disabled, variant = 'banner', aspectRatio, onChange }: { label: string; file: File | null; initialUrl?: string | null; removed: boolean; disabled?: boolean; variant?: 'logo' | 'banner'; aspectRatio?: number; onChange: (file: File | null, removed: boolean) => void }) {
  const id = useId();
  const input = useRef<HTMLInputElement>(null);
  const [preview, setPreview] = useState('');
  const [error, setError] = useState('');
  const [pending, setPending] = useState<File | null>(null);
  useEffect(() => { if (!file) { setPreview(''); return; } const url = URL.createObjectURL(file); setPreview(url); return () => URL.revokeObjectURL(url); }, [file]);
  const source = preview || (!removed ? initialUrl : '');
  return <div className={`media-field media-field-${variant}${source ? ' media-field-selected' : ''}`}>
    <label htmlFor={id}>{label}</label>
    <button type="button" className="media-field-preview" disabled={disabled} onClick={() => input.current?.click()} aria-label={`${source ? 'Change' : 'Choose'} ${variant}`} aria-describedby={`${id}-help${error ? ` ${id}-error` : ''}`}>
      {source && <img src={source} alt="" />}
      <span className="media-field-prompt"><HugeiconsIcon icon={Image01Icon} size={20} strokeWidth={1.6} aria-hidden="true" />{source ? 'Change image' : variant === 'logo' ? 'Add logo' : 'Choose banner'}</span>
    </button>
    <input hidden ref={input} id={id} type="file" accept="image/png,image/jpeg,image/webp" disabled={disabled} onChange={event => { const next = event.target.files?.[0]; if (!next) return; if (!['image/png', 'image/jpeg', 'image/webp'].includes(next.type) || next.size <= 0 || next.size > 5 * 1024 * 1024) { setError('Choose a PNG, JPEG or WebP image up to 5 MB.'); event.target.value = ''; return; } setError(''); event.target.value = ''; setPending(next); }} />
    <small id={`${id}-help`}>{file ? file.name : variant === 'logo' ? 'Square image · up to 5 MB' : 'PNG, JPEG or WebP · up to 5 MB'}</small>
    {source && <button className="quiet-link" type="button" disabled={disabled} onClick={() => onChange(null, true)}>Remove {variant}</button>}
    {removed && initialUrl && <button className="quiet-link" type="button" disabled={disabled} onClick={() => onChange(null, false)}>Keep current {variant}</button>}
    {error && <p id={`${id}-error`} role="alert" className="field-error">{error}</p>}
    {pending && <ImageCropDialog file={pending} aspect={aspectRatio || (variant === 'logo' ? 1 : 3)} label={variant} onApply={next => { onChange(next, false); setPending(null); }} onCancel={() => setPending(null)} />}
  </div>;
}
