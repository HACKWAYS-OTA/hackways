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
    default: "bg-slate-800 text-slate-300 border border-slate-700",
    success: "bg-emerald-950 text-emerald-300 border border-emerald-800",
    warning: "bg-amber-950 text-amber-300 border border-amber-800",
    danger: "bg-rose-950 text-rose-300 border border-rose-800",
    brand: "bg-indigo-950 text-indigo-300 border border-indigo-800",
  };

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold tracking-wide ${variantStyles[variant]} ${className}`}
    >
      {children}
    </span>
  );
};
