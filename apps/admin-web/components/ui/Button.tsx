import React from "react";

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline" | "ghost";
  size?: "sm" | "md" | "lg";
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
    "inline-flex items-center justify-center rounded-full text-sm font-semibold leading-5 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#34349C] focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50";

  const sizeStyles = {
    sm: "min-h-9 px-[18px] py-2",
    md: "min-h-12 px-7 py-3.5",
    lg: "min-h-12 px-8 py-3.5 text-base",
  };

  const variantStyles = {
    primary: "bg-[#34349C] text-white hover:bg-[#28287A] active:bg-[#202064]",
    secondary: "bg-[#EEF0FA] text-[#34349C] hover:bg-[#E4E7F4] active:bg-[#D9DDEE]",
    outline: "border border-[#E4E7F0] bg-white text-[#0E121E] hover:bg-[#F1F3F9] active:bg-[#E4E7F0]",
    ghost: "text-[#475069] hover:bg-[#F1F3F9] hover:text-[#0E121E] active:bg-[#E4E7F0]",
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
