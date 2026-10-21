import React, { useEffect } from "react";

export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: React.ReactNode;
}

export const Modal: React.FC<ModalProps> = ({
  isOpen,
  onClose,
  title,
  description,
  children,
}) => {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        onClose();
      }
    };

    if (isOpen) {
      document.body.style.overflow = "hidden";
      window.addEventListener("keydown", handleKeyDown);
    }

    return () => {
      document.body.style.overflow = "unset";
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#0E121E]/40 p-4 backdrop-blur-sm animate-fade-in"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
    >
      <div
        className="fixed inset-0"
        onClick={onClose}
        aria-hidden="true"
      />
      <div className="relative z-10 w-full max-w-lg overflow-hidden rounded-2xl border border-[#E4E7F0] bg-white p-6 shadow-2xl sm:p-7">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 id="modal-title" className="text-xl font-bold text-[#0E121E]">
              {title}
            </h2>
            {description && (
              <p className="text-sm text-[#475069] mt-0.5">{description}</p>
            )}
          </div>
          <button
            onClick={onClose}
            className="rounded-full p-2 text-[#757F9A] transition hover:bg-[#F1F3F9] hover:text-[#0E121E] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#34349C]"
            aria-label="Close dialog"
          >
            <svg
              className="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>
        <div>{children}</div>
      </div>
    </div>
  );
};
