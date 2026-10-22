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
  title: "Hackways Operations",
  description: "Hackways platform operations console.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <html lang="en"><body className={`${plusJakartaSans.variable} min-h-screen bg-[#f5f7fb] text-slate-950`}><UIProvider>{children}</UIProvider></body></html>;
}
