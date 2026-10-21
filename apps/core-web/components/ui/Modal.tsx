"use client";
import { useEffect, useId, useRef, type ReactNode } from 'react';
import { Icon } from './Icon';
export interface ModalProps { isOpen: boolean; onClose: () => void; title: string; description?: string; children: ReactNode; dismissible?: boolean }
export function Modal({ isOpen, onClose, title, description, children, dismissible = true }: ModalProps) {
  const dialog = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  useEffect(() => {
    const element = dialog.current;
    if (!isOpen || !element) return;
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const overflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    element.showModal();
    element.querySelector<HTMLElement>('[data-dialog-focus]')?.focus({ preventScroll: true });
    return () => { element.close(); document.body.style.overflow = overflow; previous?.focus({ preventScroll: true }); };
  }, [isOpen]);
  return <dialog ref={dialog} className="event-dialog" aria-labelledby={titleId} aria-describedby={description ? `${titleId}-description` : undefined} onCancel={(event) => { event.preventDefault(); if (dismissible) onClose(); }} onClick={(event) => { if (dismissible && event.target === event.currentTarget) onClose(); }}>
    <div className="dialog-surface"><header className="dialog-header"><div><h2 id={titleId} tabIndex={-1} data-dialog-focus>{title}</h2>{description && <p id={`${titleId}-description`}>{description}</p>}</div><button type="button" className="icon-button" aria-label="Close event details" disabled={!dismissible} onClick={onClose}><Icon name="close" /></button></header><div className="dialog-body">{children}</div></div>
  </dialog>;
}
