import React from "react";

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "inverted" | "secondary" | "outline" | "ghost";
  size?: "sm" | "md";
  isLoading?: boolean;
}

export const Button: React.FC<ButtonProps> = ({
  children,
  variant = "primary",
  size = "md",
  isLoading = false,
  className = "",
  disabled,
  ...props
}) => {
  const baseStyles =
    "inline-flex items-center justify-center font-normal rounded-full font-heading transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)] focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed select-none whitespace-nowrap";

  const sizeStyles = {
    sm: "px-[18px] py-[8px] text-[13px] leading-[16px] min-h-[36px]",
    md: "px-[28px] py-[14px] text-[14px] leading-[20px] min-h-[48px]",
  };

  const variantStyles = {
    primary:
      "bg-[var(--accent)] text-white border border-[var(--accent)] hover:bg-[var(--accent-hover)] hover:border-[var(--accent-hover)]",
    inverted:
      "bg-white text-[var(--accent-text)] border border-[var(--line)] hover:bg-white hover:border-[var(--accent)]",
    secondary:
      "bg-white text-[var(--accent-text)] border border-[var(--line)] hover:bg-white hover:border-[var(--accent)]",
    outline:
      "bg-white text-[var(--accent-text)] border border-[var(--accent)] hover:bg-white hover:border-[var(--accent-text)]",
    ghost:
      "bg-white text-[var(--accent-text)] hover:bg-white",
  };

  return (
    <button
      className={`${baseStyles} ${sizeStyles[size]} ${variantStyles[variant]} ${className}`}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading ? (
        <span className="mr-2 inline-block h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
      ) : null}
      {children}
    </button>
  );
};
