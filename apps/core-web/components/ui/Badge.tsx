import type { ReactNode } from 'react';
export interface BadgeProps { children: ReactNode; variant?: 'default' | 'success' | 'warning' | 'danger' | 'brand'; className?: string }
export function Badge({ children, variant = 'default', className = '' }: BadgeProps) {
  return <span className={`app-badge app-badge-${variant} ${className}`}>{children}</span>;
}
