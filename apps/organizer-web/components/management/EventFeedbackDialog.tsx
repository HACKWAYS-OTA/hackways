'use client';

import { useEffect, useRef } from 'react';
import { Icon } from '@/components/ui/Icon';

export function EventFeedbackDialog({ feedback, onClose }: { feedback: { kind: 'success' | 'error'; message?: string } | null; onClose: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null);
  const open = feedback !== null;
  useEffect(() => {
    const element = dialog.current;
    if (open && element && !element.open) element.showModal();
    else if (!open && element?.open) element.close();
  }, [open]);
  const failed = feedback?.kind === 'error';
  return <dialog ref={dialog} className="event-feedback" data-kind={feedback?.kind} aria-labelledby="event-feedback-title" aria-describedby={failed ? 'event-feedback-message' : undefined} onCancel={onClose} onClose={onClose} onClick={e => { if (e.target === e.currentTarget) { const bounds = e.currentTarget.getBoundingClientRect(); if (e.clientX < bounds.left || e.clientX > bounds.right || e.clientY < bounds.top || e.clientY > bounds.bottom) onClose(); } }}>
    <span className="event-feedback-icon" aria-hidden="true"><Icon name={failed ? 'alert' : 'check'} size={44} /></span>
    <h2 id="event-feedback-title">{failed ? 'Couldn’t complete your changes' : 'Changes saved'}</h2>
    {failed && <p id="event-feedback-message">{feedback?.message}</p>}
    <button type="button" className="app-button app-button-primary" onClick={onClose} autoFocus>{failed ? 'Back to editor' : 'Done'}</button>
  </dialog>;
}
