'use client';
import { useEffect, useRef, useState } from 'react';
import Image from 'next/image';
import { Icon } from '@/components/ui/Icon';
export function CoverPicker({ file, onChange, disabled, initialUrl = '', onRemoveExisting, onError }: { file: File | null; onChange: (file: File | null) => void; disabled: boolean; initialUrl?: string; onRemoveExisting?: () => void; onError?: (message: string) => void }) {
  const picker = useRef<HTMLInputElement>(null);
  const [preview, setPreview] = useState('');
  const [error, setError] = useState('');
  useEffect(() => {
    if (!file) { setPreview(''); return; }
    const url = URL.createObjectURL(file); setPreview(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);
  function choose(selected: File | undefined) {
    if (!selected) return;
    if (!['image/jpeg', 'image/png', 'image/webp'].includes(selected.type) || selected.size <= 0 || selected.size > 5 * 1024 * 1024) { if (onError) onError('Choose a JPG, PNG or WebP image up to 5 MB.'); else setError('Choose a JPG, PNG or WebP image up to 5 MB.'); return; }
    setError(''); onChange(selected);
  }
  const image = preview || initialUrl;
  return <aside className="compose-cover-column" aria-label="Event banner"><input ref={picker} type="file" accept="image/jpeg,image/png,image/webp" className="visually-hidden" aria-label="Choose event banner" tabIndex={-1} disabled={disabled} onChange={e => { choose(e.target.files?.[0]); e.target.value = ''; }} /><button type="button" className={`compose-cover-picker${image ? ' has-cover' : ''}`} disabled={disabled} onClick={() => picker.current?.click()} aria-label={image ? 'Change event banner' : 'Add event banner'}>{image ? <><Image src={image} alt="Selected event banner" fill unoptimized sizes="(max-width: 640px) 100vw, 360px" /><span className="cover-change"><Icon name="image" size={17} />Change image</span></> : <><span className="cover-upload-icon"><Icon name="image" size={27} /></span><span>Add an event banner</span><span className="cover-upload-help">Choose an image for your event</span></>}</button><div className="compose-cover-caption"><p>JPG, PNG or WebP. Up to 5 MB.</p>{(file || initialUrl) && <button type="button" className="quiet-link" disabled={disabled} onClick={() => { if (file) onChange(null); else onRemoveExisting?.(); setError(''); }}>Remove image</button>}</div>{file && <p className="cover-file-name">{file.name}</p>}{error && <p className="journey-error" role="alert">{error}</p>}</aside>;
}
