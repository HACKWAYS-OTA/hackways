import type { Metadata } from "next";
import "./globals.css";
import React from "react";

export const metadata: Metadata = {
  title: "Hackways — High-Concurrency Tech Event Platform",
  description: "Concurrency-safe developer event registration, real-time ticket tier inventory, and zero-overbooking guarantees.",
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
                  Hackways
                </span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-slate-800 text-slate-400 font-mono">
                  hackways.in
                </span>
              </a>
              <nav className="hidden md:flex items-center gap-4 text-sm font-medium text-slate-300">
                <a href="#features" className="hover:text-white transition">
                  Features
                </a>
                <a href="#architecture" className="hover:text-white transition">
                  Architecture
                </a>
                <a href="#events" className="hover:text-white transition">
                  Live Catalog
                </a>
              </nav>
            </div>
            <div className="flex items-center gap-3">
              <a
                href="http://localhost:3002"
                className="text-sm font-medium text-slate-300 hover:text-white px-3 py-1.5 rounded-lg hover:bg-slate-800 transition"
              >
                Organizer Studio
              </a>
              <a
                href="http://localhost:3001"
                className="text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg transition shadow-sm"
              >
                Attendee Portal
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
