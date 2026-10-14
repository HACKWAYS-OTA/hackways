import type { Metadata } from "next";
import "./globals.css";
import React from "react";

export const metadata: Metadata = {
  title: "Hackways Attendee Portal — Discover, Host & RSVP to Tech Events",
  description: "Attendee event discovery, ticket selection, and concurrency-safe RSVP platform.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="bg-slate-950 text-slate-100 flex flex-col min-h-screen">
        <header className="border-b border-slate-800 bg-slate-900/60 backdrop-blur sticky top-0 z-50">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <div className="flex items-center gap-6">
              <a href="/" className="flex items-center gap-2">
                <span className="w-8 h-8 rounded-lg bg-indigo-600 flex items-center justify-center font-bold text-white text-lg shadow-sm">
                  H
                </span>
                <span className="font-bold text-xl tracking-tight text-white">
                  Hackways <span className="text-indigo-400 font-normal">App</span>
                </span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-slate-800 text-slate-400 font-mono">
                  app.hackways.in
                </span>
              </a>
              <nav className="hidden md:flex items-center gap-4 text-sm font-medium text-slate-300">
                <a href="#events" className="hover:text-white transition">
                  Events
                </a>
                <a href={process.env.NEXT_PUBLIC_PUBLIC_URL || "https://hackways.in"} className="hover:text-white transition">
                  About
                </a>
              </nav>
            </div>
            <div className="flex items-center gap-3">
              <a
                href={process.env.NEXT_PUBLIC_ORGANIZER_URL || "https://organizer.hackways.in"}
                className="text-sm font-medium text-slate-300 hover:text-white px-3 py-1.5 rounded-lg hover:bg-slate-800 transition"
              >
                Host an Event
              </a>
              <a
                href="#signin"
                className="text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg transition shadow-sm"
              >
                Sign In
              </a>
            </div>
          </div>
        </header>

        <main className="flex-1">{children}</main>

        <footer className="border-t border-slate-800 bg-slate-900 py-8 text-center text-sm text-slate-500">
          <div className="max-w-7xl mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-4">
            <p>© {new Date().getFullYear()} Hackways Platform. All rights reserved.</p>
            <div className="flex items-center gap-4 text-slate-400">
              <a href="/api/v1" className="hover:text-white transition">
                API Docs
              </a>
              <span>•</span>
              <span className="text-emerald-400 flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                System Operational
              </span>
            </div>
          </div>
        </footer>
      </body>
    </html>
  );
}
