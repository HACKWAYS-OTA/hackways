'use client';
import { useEffect, useId, useRef, useState, type PointerEvent } from 'react';
import { createPortal } from 'react-dom';
import { HugeiconsIcon } from '@hugeicons/react';
import Cancel01Icon from '@hugeicons/core-free-icons/Cancel01Icon';
const clamp = (value: number) => Math.max(0, Math.min(100, value));

export function ImageCropDialog({ file, aspect, label, onApply, onCancel }: { file: File; aspect: number; label: string; onApply: (file: File) => void; onCancel: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null), canvas = useRef<HTMLCanvasElement>(null), heading = useRef<HTMLHeadingElement>(null);
  const id = useId();
  const [image, setImage] = useState<HTMLImageElement | null>(null), [error, setError] = useState(''), [busy, setBusy] = useState(false);
  const [zoom, setZoom] = useState(1), [x, setX] = useState(50), [y, setY] = useState(50);
  const drag = useRef<{ clientX: number; clientY: number; x: number; y: number }>();
  const alive = useRef(true);
  useEffect(() => {
    alive.current = true;
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const element = dialog.current; element?.showModal(); heading.current?.focus();
    return () => { alive.current = false; element?.close(); previous?.focus(); };
  }, []);
  useEffect(() => {
    let cancelled = false; const url = URL.createObjectURL(file), next = new Image();
    next.onload = () => { if (!cancelled) { if (!next.naturalWidth || !next.naturalHeight || next.naturalWidth * next.naturalHeight > 40000000) setError('Choose an image smaller than 40 megapixels.'); else setImage(next); } };
    next.onerror = () => { if (!cancelled) setError('This image could not be opened. Choose another PNG, JPEG or WebP.'); }; next.src = url;
    return () => { cancelled = true; URL.revokeObjectURL(url); };
  }, [file]);
  const cropWidth = image ? Math.min(image.naturalWidth, image.naturalHeight * aspect) / zoom : 0;
  const cropHeight = cropWidth / aspect;
  useEffect(() => {
    const target = canvas.current, context = target?.getContext('2d');
    if (!image || !target || !context) return;
    context.clearRect(0, 0, target.width, target.height);
    context.imageSmoothingEnabled = true; context.imageSmoothingQuality = 'high';
    context.drawImage(image, (image.naturalWidth - cropWidth) * x / 100, (image.naturalHeight - cropHeight) * y / 100, cropWidth, cropHeight, 0, 0, target.width, target.height);
  }, [image, cropWidth, cropHeight, x, y]);
  function move(event: PointerEvent<HTMLCanvasElement>) {
    if (!drag.current || !image) return;
    const rect = event.currentTarget.getBoundingClientRect();
    if (image.naturalWidth > cropWidth) setX(clamp(drag.current.x - (event.clientX - drag.current.clientX) / rect.width * cropWidth / (image.naturalWidth - cropWidth) * 100));
    if (image.naturalHeight > cropHeight) setY(clamp(drag.current.y - (event.clientY - drag.current.clientY) / rect.height * cropHeight / (image.naturalHeight - cropHeight) * 100));
  }
  async function apply() {
    if (!image || busy) return; setBusy(true); setError('');
    try {
      const output = document.createElement('canvas'); output.width = Math.max(1, Math.min(aspect === 1 ? 768 : 1600, Math.round(cropWidth))); output.height = Math.max(1, Math.round(output.width / aspect));
      const context = output.getContext('2d'); if (!context) throw new Error('Image adjustment is unavailable in this browser.');
      context.imageSmoothingEnabled = true; context.imageSmoothingQuality = 'high';
      context.drawImage(image, (image.naturalWidth - cropWidth) * x / 100, (image.naturalHeight - cropHeight) * y / 100, cropWidth, cropHeight, 0, 0, output.width, output.height);
      const blob = await new Promise<Blob>((resolve, reject) => output.toBlob(result => result ? resolve(result) : reject(new Error('The adjusted image could not be saved.')), file.type, .95));
      if (blob.size > 5 * 1024 * 1024) throw new Error('The adjusted image is too large. Choose a smaller image.');
      const extension = blob.type === 'image/jpeg' ? 'jpg' : blob.type === 'image/webp' ? 'webp' : 'png';
      if (alive.current) onApply(new File([blob], `${file.name.replace(/\.[^.]+$/, '')}-cropped.${extension}`, { type: blob.type }));
    } catch (failure) { if (alive.current) setError(failure instanceof Error ? failure.message : 'The adjusted image could not be saved.'); }
    finally { if (alive.current) setBusy(false); }
  }
  return createPortal(<dialog className="image-crop-dialog" ref={dialog} aria-labelledby={`${id}-title`} aria-describedby={`${id}-description`} onCancel={event => { event.preventDefault(); onCancel(); }}>
    <header><div><h2 id={`${id}-title`} ref={heading} tabIndex={-1}>Adjust your {label}</h2><p id={`${id}-description`}>Drag the image or use the controls to find the right fit.</p></div><button className="image-crop-close" type="button" onClick={onCancel} disabled={busy} aria-label="Cancel image adjustment"><HugeiconsIcon icon={Cancel01Icon} size={20} strokeWidth={1.8} /></button></header>
    <div className={`image-crop-frame${aspect === 1 ? ' image-crop-square' : ''}`} style={{ aspectRatio: aspect }}>
      {image ? <canvas ref={canvas} width={1000} height={Math.round(1000 / aspect)} role="img" aria-label={`Adjusted ${label} preview`} onPointerDown={event => { event.currentTarget.setPointerCapture(event.pointerId); drag.current = { clientX: event.clientX, clientY: event.clientY, x, y }; }} onPointerMove={move} onPointerUp={() => { drag.current = undefined; }} onPointerCancel={() => { drag.current = undefined; }} /> : <p role="status">{error ? 'Image unavailable' : 'Preparing your image…'}</p>}
    </div>
    <div className="image-crop-controls"><label htmlFor={`${id}-zoom`}>Zoom<span>{Math.round(zoom * 100)}%</span><input id={`${id}-zoom`} type="range" min="1" max="3" step="0.01" value={zoom} onChange={event => setZoom(Number(event.target.value))} disabled={!image || busy} /></label><label htmlFor={`${id}-horizontal`}>Horizontal position<input id={`${id}-horizontal`} type="range" min="0" max="100" value={x} onChange={event => setX(Number(event.target.value))} disabled={!image || busy || image.naturalWidth <= cropWidth} /></label><label htmlFor={`${id}-vertical`}>Vertical position<input id={`${id}-vertical`} type="range" min="0" max="100" value={y} onChange={event => setY(Number(event.target.value))} disabled={!image || busy || image.naturalHeight <= cropHeight} /></label></div>
    {error && <p className="image-crop-error" role="alert">{error}</p>}
    <footer><button className="image-crop-reset" type="button" disabled={!image || busy} onClick={() => { setZoom(1); setX(50); setY(50); }}>Reset</button><button className="app-button app-button-outline" type="button" onClick={onCancel} disabled={busy}>Cancel</button><button className="app-button app-button-primary" type="button" onClick={() => void apply()} disabled={!image || busy}>{busy ? 'Saving…' : 'Use image'}</button></footer>
  </dialog>, document.body);
}
