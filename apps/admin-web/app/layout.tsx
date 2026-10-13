import type { Metadata } from "next";
import "./globals.css";
import React from "react";

export const metadata: Metadata = {
  title: "Hackways Admin Portal — Platform Management & Auditing",
  description: "Network-isolated administration dashboard for platform health, event moderation, and telemetry.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="bg-slate-950 text-slate-100 flex flex-col min-h-screen">
        <header className="border-b border-rose-900/50 bg-slate-900/80 backdrop-blur sticky top-0 z-50">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <div className="flex items-center gap-6">
              <a href="/" className="flex items-center gap-2">
                <span className="w-8 h-8 rounded-lg bg-rose-600 flex items-center justify-center font-bold text-white text-lg shadow-sm">
                  A
                </span>
                <span className="font-bold text-xl tracking-tight text-white">
                  Hackways <span className="text-rose-400 font-normal">Admin</span>
                </span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-rose-950 border border-rose-800 text-rose-300 font-mono">
                  admin.hackways.in [VPN ISOLATED]
                </span>
              </a>
              <nav className="hidden md:flex items-center gap-4 text-sm font-medium text-slate-300">
                <a href="#health" className="hover:text-white transition">
                  System Health
                </a>
                <a href="#moderation" className="hover:text-white transition">
                  Event Moderation
                </a>
                <a href="#audit" className="hover:text-white transition">
                  Audit Logs
                </a>
              </nav>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-xs text-rose-400 font-mono flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-rose-500 animate-pulse" />
                IP: Restricted
              </span>
              <a
                href="#logout"
                className="text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-300 px-3 py-1.5 rounded-lg transition"
              >
                Sign Out
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
