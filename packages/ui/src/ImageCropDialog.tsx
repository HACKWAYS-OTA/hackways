'use client';
import { useEffect, useId, useRef, useState, type CSSProperties, type PointerEvent } from 'react';
import { createPortal } from 'react-dom';
import { HugeiconsIcon } from '@hugeicons/react';
import Cancel01Icon from '@hugeicons/core-free-icons/Cancel01Icon';
import ZoomInIcon from '@hugeicons/core-free-icons/ZoomInIcon';
import ZoomOutIcon from '@hugeicons/core-free-icons/ZoomOutIcon';
import ArrowRight01Icon from '@hugeicons/core-free-icons/ArrowRight01Icon';
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
    const overflow = document.body.style.overflow; document.body.style.overflow = 'hidden';
    const element = dialog.current; element?.showModal(); heading.current?.focus();
    return () => { alive.current = false; element?.close(); document.body.style.overflow = overflow; previous?.focus(); };
  }, []);
  useEffect(() => {
    setImage(null); setError(''); setZoom(1); setX(50); setY(50);
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
    if (!drag.current || !image || busy) return;
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
  function changeZoom(value: number) { setZoom(Math.max(1, Math.min(3, Math.round(value * 100) / 100))); }
  return createPortal(<dialog className="image-crop-dialog" ref={dialog} aria-labelledby={`${id}-title`} aria-describedby={`${id}-description`} onCancel={event => { event.preventDefault(); if (!busy) onCancel(); }}>
    <header><div><h2 id={`${id}-title`} ref={heading} tabIndex={-1}>Adjust {label}</h2><p id={`${id}-description`}>Drag to reposition. Zoom to get the right fit.</p></div><button className="image-crop-close" type="button" onClick={onCancel} disabled={busy} aria-label="Cancel image adjustment"><HugeiconsIcon icon={Cancel01Icon} size={20} strokeWidth={1.8} /></button></header>
    <div className="image-crop-body">
      <div className="image-crop-stage"><div className="image-crop-frame" style={{ '--crop-aspect': aspect } as CSSProperties}>
        {image ? <canvas ref={canvas} width={1000} height={Math.round(1000 / aspect)} role="img" tabIndex={busy ? -1 : 0} aria-label={`Adjust ${label} position`} aria-describedby={`${id}-keyboard`} aria-keyshortcuts="ArrowUp ArrowDown ArrowLeft ArrowRight" onKeyDown={event => {
          if (busy || !['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(event.key)) return;
          event.preventDefault(); const step = event.shiftKey ? 10 : 2;
          if (event.key === 'ArrowLeft') setX(value => clamp(value + step));
          if (event.key === 'ArrowRight') setX(value => clamp(value - step));
          if (event.key === 'ArrowUp') setY(value => clamp(value + step));
          if (event.key === 'ArrowDown') setY(value => clamp(value - step));
        }} onPointerDown={event => { if (busy) return; event.currentTarget.setPointerCapture(event.pointerId); drag.current = { clientX: event.clientX, clientY: event.clientY, x, y }; }} onPointerMove={move} onPointerUp={() => { drag.current = undefined; }} onPointerCancel={() => { drag.current = undefined; }} onLostPointerCapture={() => { drag.current = undefined; }} /> : <p role="status">{error ? 'Image unavailable' : 'Preparing your image...'}</p>}
        <div className="image-crop-guides" aria-hidden="true"><span /><span /><span /><span /></div>
      </div></div>
      <p className="image-crop-sr-only" id={`${id}-keyboard`}>Use the arrow keys to reposition the image. Hold Shift for larger steps.</p>
      <div className="image-crop-tools">
      <div className="image-crop-toolbar"><label htmlFor={`${id}-zoom`}>Zoom</label><div className="image-crop-zoom"><button type="button" aria-label="Zoom out" onClick={() => changeZoom(zoom - .1)} disabled={!image || busy || zoom <= 1}><HugeiconsIcon icon={ZoomOutIcon} size={20} strokeWidth={1.6} /></button><input id={`${id}-zoom`} type="range" min="1" max="3" step="0.01" value={zoom} aria-valuetext={`${Math.round(zoom * 100)} percent`} onChange={event => changeZoom(Number(event.target.value))} disabled={!image || busy} /><button type="button" aria-label="Zoom in" onClick={() => changeZoom(zoom + .1)} disabled={!image || busy || zoom >= 3}><HugeiconsIcon icon={ZoomInIcon} size={20} strokeWidth={1.6} /></button></div><output htmlFor={`${id}-zoom`}>{Math.round(zoom * 100)}%</output></div>
      <div className="image-crop-pan" role="group" aria-label="Reposition image"><span>Position</span>{(['left', 'right', 'up', 'down'] as const).map(direction => <button key={direction} type="button" aria-label={`Move image ${direction}`} disabled={!image || busy || (direction === 'left' || direction === 'right' ? image.naturalWidth <= cropWidth || (direction === 'left' ? x >= 100 : x <= 0) : image.naturalHeight <= cropHeight || (direction === 'up' ? y >= 100 : y <= 0))} onClick={() => { if (direction === 'left') setX(value => clamp(value + 3)); if (direction === 'right') setX(value => clamp(value - 3)); if (direction === 'up') setY(value => clamp(value + 3)); if (direction === 'down') setY(value => clamp(value - 3)); }}><HugeiconsIcon icon={ArrowRight01Icon} className={`image-crop-pan-${direction}`} size={18} strokeWidth={1.8} /></button>)}</div>
      </div>
      {error && <p className="image-crop-error" role="alert">{error}</p>}
    </div>
    <footer><button className="image-crop-reset" type="button" disabled={!image || busy} onClick={() => { setZoom(1); setX(50); setY(50); }}>Reset</button><button className="app-button app-button-outline" type="button" onClick={onCancel} disabled={busy}>Cancel</button><button className="app-button app-button-primary" type="button" onClick={() => void apply()} disabled={!image || busy}>{busy ? 'Saving...' : 'Use image'}</button></footer>
  </dialog>, document.body);
}
