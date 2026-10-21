'use client';
import { useEffect, useRef, type ReactNode } from 'react';
import { Icon } from '@/components/ui/Icon';

export function EditorMenu({ label, children, disabled = false }: { label: ReactNode; children: ReactNode; disabled?: boolean }) {
  const ref = useRef<HTMLDetailsElement>(null);
  useEffect(() => {
    function close(e: PointerEvent) { if (e.target instanceof Node && !ref.current?.contains(e.target)) ref.current?.removeAttribute('open'); }
    function escape(e: KeyboardEvent) { if (e.key === 'Escape' && ref.current?.open) { ref.current.open = false; ref.current.querySelector('summary')?.focus(); } }
    document.addEventListener('pointerdown', close); document.addEventListener('keydown', escape);
    return () => { document.removeEventListener('pointerdown', close); document.removeEventListener('keydown', escape); };
  }, []);
  return <details ref={ref} className="editor-menu" onClick={e => { if (e.target instanceof Element && e.target.closest('button,a')) ref.current?.removeAttribute('open'); }}><summary onClick={e => { if (disabled) e.preventDefault(); }} aria-disabled={disabled || undefined}>{label}<Icon name="down" size={14} /></summary><div className="editor-menu-panel">{children}</div></details>;
}
