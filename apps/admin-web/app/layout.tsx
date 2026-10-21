import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import type { Metadata } from "next";
import { Plus_Jakarta_Sans } from "next/font/google";
import "./globals.css";
import React from "react";

const plusJakartaSans = Plus_Jakarta_Sans({
  subsets: ["latin"],
  variable: "--font-plus-jakarta-sans",
});

export const metadata: Metadata = {
  title: "Hackways | Platform Operations",
  description: "Backend readiness information for Hackways platform operators.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={`${plusJakartaSans.variable} flex min-h-screen flex-col bg-[#F8F9FD] text-[#0E121E]`}><UIProvider>
        <header className="sticky top-0 z-50 border-b border-[#E4E7F0] bg-white/95 backdrop-blur">
          <div className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
            <div className="flex items-center gap-3">
              <a href="/" className="flex items-center gap-2">
                <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-[#34349C] text-lg font-bold text-white">
                  A
                </span>
                <span className="text-xl font-bold tracking-tight text-[#0E121E]">
                  Hackways <span className="font-normal text-[#34349C]">Operations</span>
                </span>
                <span className="hidden rounded-full bg-[#F1F3F9] px-2 py-0.5 font-mono text-xs text-[#757F9A] lg:inline">
                  admin.hackways.in
                </span>
              </a>
              <nav className="hidden items-center gap-4 text-sm font-medium text-[#475069] md:flex">
                <a href="/" className="transition hover:text-[#34349C]">
                  System Health
                </a>
              </nav>
            </div>
            <a href={process.env.NEXT_PUBLIC_PUBLIC_URL || "https://hackways.in"} className="rounded-full px-3 py-3 text-sm font-semibold text-[#34349C] transition hover:bg-[#EEF0FA]">Public site</a>
          </div>
        </header>

        <main className="flex-1">{children}</main>

        <footer className="border-t border-[#E4E7F0] bg-white py-6 text-center text-sm text-[#757F9A]">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <p>Copyright {new Date().getFullYear()} Hackways Platform.</p>
          </div>
        </footer>
      </UIProvider></body>
    </html>
  );
}
