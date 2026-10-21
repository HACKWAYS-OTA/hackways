'use client';
import type { ButtonHTMLAttributes } from 'react';

/** Move focus and scroll without creating a URL fragment or a history entry. */
export function scrollToSection(id: string, behavior: ScrollBehavior = 'auto'): boolean {
  const target = document.getElementById(id);
  if (!target) return false;
  if (!target.hasAttribute('tabindex')) target.tabIndex = -1;
  target.focus({ preventScroll: true });
  target.scrollIntoView({ block: 'start', behavior });
  return true;
}
export function SectionJumpButton({ targetId, onClick, className = '', ...props }: Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'type'> & { targetId: string }) {
  return <button {...props} type="button" className={`hw-section-jump ${className}`} onClick={event => { onClick?.(event); if (!event.defaultPrevented) scrollToSection(targetId, window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth'); }} />;
}
