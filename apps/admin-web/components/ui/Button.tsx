import type { ButtonHTMLAttributes } from 'react';
export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
}
export function Button({ children, variant = 'primary', size = 'md', isLoading = false, className = '', disabled, type = 'button', ...props }: ButtonProps) {
  return <button {...props} type={type} className={`app-button app-button-${variant} app-button-${size} ${className}`} disabled={disabled || isLoading} aria-busy={isLoading || undefined}>{isLoading && <span className="button-spinner" aria-hidden="true" />}{children}</button>;
}
