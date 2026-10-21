import React from "react";

export interface BadgeProps {
  children: React.ReactNode;
  variant?: "default" | "success" | "warning" | "danger" | "brand";
  className?: string;
}

export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = "default",
  className = "",
}) => {
  const variantStyles = {
    default: "text-[#475069]",
    success: "text-emerald-800",
    warning: "text-amber-800",
    danger: "text-rose-800",
    brand: "text-[#34349C]",
  };

  return (
    <span
      className={`inline-flex items-center text-xs font-medium ${variantStyles[variant]} ${className}`}
    >
      {children}
    </span>
  );
};
