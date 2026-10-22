"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

type HealthSnapshot = {
  status: string;
  database?: string;
  version?: string;
  uptime?: string;
  environment?: string;
};

type ServiceState = { label: string; data: HealthSnapshot | null; error: string | null };
type NavItem = { label: string; icon: string; active?: boolean };

const navItems: NavItem[] = [
  { label: "Overview", icon: "grid", active: true },
  { label: "Events & spaces", icon: "calendar" },
  { label: "People & access", icon: "users" },
  { label: "Moderation", icon: "shield" },
  { label: "Messaging", icon: "mail" },
  { label: "Storage", icon: "cloud" },
  { label: "Audit log", icon: "list" },
];

function Icon({ name, size = 18 }: { name: string; size?: number }) {
  const common = { width: size, height: size, viewBox: "0 0 24 24", fill: "none", stroke: "currentColor", strokeWidth: 1.8, strokeLinecap: "round" as const, strokeLinejoin: "round" as const, "aria-hidden": true };
  switch (name) {
    case "grid": return <svg {...common}><rect x="3" y="3" width="7" height="7" rx="1" /><rect x="14" y="3" width="7" height="7" rx="1" /><rect x="3" y="14" width="7" height="7" rx="1" /><rect x="14" y="14" width="7" height="7" rx="1" /></svg>;
    case "calendar": return <svg {...common}><rect x="3" y="4.5" width="18" height="17" rx="2" /><path d="M16 2.5v4M8 2.5v4M3 9.5h18" /><path d="M8 14h.01M12 14h.01M16 14h.01M8 18h.01M12 18h.01" /></svg>;
    case "users": return <svg {...common}><path d="M16 20v-1.5a4 4 0 0 0-4-4H7a4 4 0 0 0-4 4V20" /><circle cx="9.5" cy="7" r="3.5" /><path d="M17 11a3.5 3.5 0 0 0 0-7M21 20v-1.5a4 4 0 0 0-3-3.87" /></svg>;
    case "shield": return <svg {...common}><path d="M12 3 20 6v5c0 5-3.4 8.5-8 10-4.6-1.5-8-5-8-10V6l8-3Z" /><path d="m8.5 12 2.2 2.2 4.8-4.8" /></svg>;
    case "mail": return <svg {...common}><rect x="3" y="5" width="18" height="14" rx="2" /><path d="m3 7 9 6 9-6" /></svg>;
    case "cloud": return <svg {...common}><path d="M7 18h10.5a4.5 4.5 0 0 0 .7-8.95A6.5 6.5 0 0 0 5.1 10.7 3.7 3.7 0 0 0 7 18Z" /></svg>;
    case "list": return <svg {...common}><path d="M8 6h13M8 12h13M8 18h13" /><path d="M3 6h.01M3 12h.01M3 18h.01" /></svg>;
    case "external": return <svg {...common}><path d="M14 4h6v6M20 4l-9 9" /><path d="M18 13v5a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h5" /></svg>;
    case "refresh": return <svg {...common}><path d="M20 11a8 8 0 0 0-14.7-3L3 10" /><path d="M3 4v6h6M4 13a8 8 0 0 0 14.7 3L21 14" /><path d="M21 20v-6h-6" /></svg>;
    default: return null;
  }
}

async function readHealth(path: string): Promise<HealthSnapshot> {
  const response = await fetch(path, { cache: "no-store" });
  const body = (await response.json()) as HealthSnapshot;
  if (!response.ok && !body.status) throw new Error(`Health endpoint returned HTTP ${response.status}.`);
  return body;
}

function StatusDot({ healthy, pending = false }: { healthy: boolean; pending?: boolean }) {
  const label = pending ? "Checking" : healthy ? "Healthy" : "Unavailable";
  return <span className={`inline-block h-2.5 w-2.5 rounded-full ${pending ? "bg-slate-300" : healthy ? "bg-emerald-500" : "bg-rose-500"}`} aria-label={label} />;
}

function StatTile({ label, value, detail, tone = "plain" }: { label: string; value: string; detail: string; tone?: "plain" | "good" | "warn" }) {
  const toneClass = tone === "good" ? "text-emerald-700" : tone === "warn" ? "text-amber-700" : "text-slate-950";
  return <div className="rounded-[14px] border border-slate-200 bg-white px-5 py-4 shadow-[0_1px_2px_rgba(15,23,42,0.03)]"><p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">{label}</p><p className={`mt-2 text-[25px] font-semibold tracking-[-0.04em] ${toneClass}`}>{value}</p><p className="mt-1 text-xs text-slate-500">{detail}</p></div>;
}

export default function AdminPortalPage() {
  const [services, setServices] = useState<{ core: ServiceState; global: ServiceState }>({ core: { label: "Core backend", data: null, error: null }, global: { label: "Global backend", data: null, error: null } });
  const [isLoading, setIsLoading] = useState(true);
  const [lastChecked, setLastChecked] = useState<Date | null>(null);

  const loadHealth = useCallback(async () => {
    setIsLoading(true);
    const [core, global] = await Promise.allSettled([readHealth("/health/core"), readHealth("/health/global")]);
    setServices({ core: core.status === "fulfilled" ? { label: "Core backend", data: core.value, error: null } : { label: "Core backend", data: null, error: "Unavailable" }, global: global.status === "fulfilled" ? { label: "Global backend", data: global.value, error: null } : { label: "Global backend", data: null, error: "Unavailable" } });
    setLastChecked(new Date());
    setIsLoading(false);
  }, []);

  useEffect(() => { void loadHealth(); }, [loadHealth]);

  const healthyCount = useMemo(() => [services.core, services.global].filter((service) => service.data?.status === "ok" && service.data.database === "healthy").length, [services]);
  const platformHealthy = !isLoading && healthyCount === 2;
  const platformState = isLoading ? "Checking systems" : platformHealthy ? "All systems operational" : "Attention needed";
  const areas = [{ icon: "calendar", label: "Events & spaces", detail: "Lifecycle, visibility, ownership, and capacity." }, { icon: "users", label: "People & access", detail: "Accounts, organizer access, and session review." }, { icon: "shield", label: "Moderation", detail: "Reports, policy actions, and platform safeguards." }, { icon: "mail", label: "Messaging", detail: "Email outbox delivery, retries, and templates." }, { icon: "cloud", label: "Storage", detail: "Media uploads, SAS activity, and retention." }, { icon: "list", label: "Audit log", detail: "A durable record of sensitive operator actions." }];

  return <div className="min-h-screen bg-[#f5f7fb] text-slate-950"><div className="mx-auto flex min-h-screen max-w-[1600px]">
    <aside className="hidden w-[244px] shrink-0 flex-col bg-[#17202e] px-4 py-5 text-slate-300 lg:flex"><a href="/" className="flex items-center gap-3 px-3"><span className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-[#5d62d6] text-sm font-bold text-white">H</span><span className="text-[15px] font-semibold tracking-[-0.02em] text-white">Hackways <span className="font-normal text-slate-400">Ops</span></span></a><div className="mt-10 px-3 text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Workspace</div><nav className="mt-3 space-y-1" aria-label="Admin sections">{navItems.map((item) => <button key={item.label} type="button" disabled={!item.active} className={`group flex w-full items-center gap-3 rounded-[9px] px-3 py-2.5 text-left text-[13px] font-medium transition ${item.active ? "bg-[#2b3446] text-white" : "text-slate-400 hover:bg-[#222c3d] hover:text-slate-200 disabled:cursor-default"}`}><span className={item.active ? "text-[#aeb1ff]" : "text-slate-500"}><Icon name={item.icon} size={17} /></span>{item.label}{!item.active && <span className="ml-auto text-[10px] text-slate-600">Soon</span>}</button>)}</nav><div className="mt-auto border-t border-slate-700/60 pt-4"><div className="flex items-center gap-3 rounded-[10px] px-3 py-2.5"><span className="flex h-8 w-8 items-center justify-center rounded-full bg-[#303b50] text-xs font-semibold text-slate-200">A</span><div><p className="text-xs font-semibold text-slate-200">Platform admin</p><p className="mt-0.5 text-[11px] text-slate-500">Network protected</p></div></div></div></aside>
    <main className="min-w-0 flex-1"><header className="flex h-[70px] items-center justify-between border-b border-slate-200 bg-white px-5 sm:px-8"><div className="flex items-center gap-3"><span className="text-sm font-medium text-slate-500">Control plane</span><span className="text-slate-300">/</span><span className="text-sm font-semibold text-slate-900">Overview</span></div><div className="flex items-center gap-3"><span className="hidden text-xs text-slate-500 sm:inline">{lastChecked ? `Updated ${lastChecked.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}` : "Checking system"}</span><button type="button" onClick={() => void loadHealth()} className="inline-flex h-9 items-center gap-2 rounded-[8px] border border-slate-200 bg-white px-3 text-xs font-semibold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 disabled:opacity-50" disabled={isLoading}><Icon name="refresh" size={14} /> Refresh</button><a href={process.env.NEXT_PUBLIC_PUBLIC_URL || "https://hackways.in"} className="hidden items-center gap-1 text-xs font-semibold text-[#4d52bb] sm:inline-flex">Public site <Icon name="external" size={13} /></a></div></header>
      <div className="mx-auto max-w-[1280px] px-5 py-7 sm:px-8 sm:py-9"><section className="flex flex-col justify-between gap-5 md:flex-row md:items-end"><div><h1 className="text-[30px] font-semibold tracking-[-0.05em] text-slate-950 sm:text-[36px]">Platform overview</h1><p className="mt-2 max-w-xl text-sm leading-6 text-slate-500">A focused view of the systems that keep Hackways events, accounts, and messaging running.</p></div><div className={`inline-flex w-fit items-center gap-2 rounded-full border px-3.5 py-2 text-xs font-semibold ${isLoading ? "border-slate-200 bg-slate-50 text-slate-600" : platformHealthy ? "border-emerald-200 bg-emerald-50 text-emerald-800" : "border-amber-200 bg-amber-50 text-amber-800"}`}><StatusDot healthy={platformHealthy} pending={isLoading} />{platformState}</div></section>
        <section className="mt-7 grid gap-3 sm:grid-cols-2 xl:grid-cols-4" aria-label="Platform summary"><StatTile label="Platform status" value={isLoading ? "Checking" : platformHealthy ? "Operational" : "Degraded"} detail="Core and global services" tone={isLoading ? "plain" : platformHealthy ? "good" : "warn"} /><StatTile label="Database" value={isLoading ? "Checking" : healthyCount === 2 ? "Connected" : `${healthyCount}/2 ready`} detail="PostgreSQL readiness" tone={isLoading ? "plain" : healthyCount === 2 ? "good" : "warn"} /><StatTile label="Service surfaces" value="6" detail="Four web apps Ã‚Â· two engines" /><StatTile label="Environment" value={services.core.data?.environment || services.global.data?.environment || "Checking"} detail={`Release ${services.core.data?.version || "Ã¢â‚¬â€"}`} /></section>
        <section className="mt-8 grid gap-5 xl:grid-cols-[minmax(0,1.45fr)_minmax(310px,0.8fr)]"><div className="rounded-[14px] border border-slate-200 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.03)]"><div className="flex items-center justify-between border-b border-slate-100 px-5 py-4"><div><h2 className="text-[15px] font-semibold text-slate-950">System pulse</h2><p className="mt-1 text-xs text-slate-500">Live readiness from the production engines.</p></div><span className="text-[11px] font-medium text-slate-400">Auto checked</span></div><div className="divide-y divide-slate-100">{[services.core, services.global].map((service) => { const healthy = service.data?.status === "ok" && service.data.database === "healthy"; return <div key={service.label} className="grid gap-3 px-5 py-4 sm:grid-cols-[1.3fr_0.7fr_0.7fr_0.7fr] sm:items-center"><div className="flex items-center gap-3"><StatusDot healthy={healthy} pending={isLoading} /><div><p className="text-sm font-semibold text-slate-900">{service.label}</p><p className="mt-0.5 text-xs text-slate-500">{service.data?.database === "healthy" ? "Database connected" : service.error || "Checking readiness"}</p></div></div><div><p className="text-[10px] uppercase tracking-[0.1em] text-slate-400">Status</p><p className={`mt-1 text-xs font-semibold ${healthy ? "text-emerald-700" : "text-amber-700"}`}>{isLoading ? "Checking" : healthy ? "Ready" : "Review"}</p></div><div><p className="text-[10px] uppercase tracking-[0.1em] text-slate-400">Uptime</p><p className="mt-1 text-xs font-medium text-slate-700">{service.data?.uptime || "Ã¢â‚¬â€"}</p></div><div><p className="text-[10px] uppercase tracking-[0.1em] text-slate-400">Version</p><p className="mt-1 font-mono text-[11px] text-slate-600">{service.data?.version || "Ã¢â‚¬â€"}</p></div></div>; })}</div></div><div className="rounded-[14px] border border-slate-200 bg-[#202a3a] px-5 py-5 text-white shadow-[0_1px_2px_rgba(15,23,42,0.03)]"><div className="flex items-center justify-between"><h2 className="text-[15px] font-semibold">Admin workspace</h2><span className="rounded-full bg-white/10 px-2 py-1 text-[10px] font-semibold text-slate-300">Roadmap</span></div><p className="mt-2 text-xs leading-5 text-slate-400">The control plane is being expanded around real, permissioned platform data.</p><div className="mt-5 space-y-3"><div className="flex items-start gap-3"><span className="mt-0.5 text-[#aeb1ff]"><Icon name="shield" size={16} /></span><div><p className="text-xs font-semibold text-slate-200">Access boundaries</p><p className="mt-1 text-[11px] leading-4 text-slate-500">Network isolation remains enabled for this console.</p></div></div><div className="flex items-start gap-3"><span className="mt-0.5 text-[#aeb1ff]"><Icon name="list" size={16} /></span><div><p className="text-xs font-semibold text-slate-200">Audit-ready operations</p><p className="mt-1 text-[11px] leading-4 text-slate-500">User, event, moderation, and delivery actions will appear here as their protected APIs land.</p></div></div></div></div></section>
        <section className="mt-8"><div className="flex items-end justify-between"><div><h2 className="text-[15px] font-semibold text-slate-950">Platform areas</h2><p className="mt-1 text-xs text-slate-500">The operational surfaces this console is designed to hold.</p></div><span className="text-xs text-slate-400">6 areas</span></div><div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">{areas.map((area) => <button key={area.label} type="button" disabled className="group flex min-h-[112px] items-start gap-3 rounded-[14px] border border-slate-200 bg-white p-4 text-left opacity-90 shadow-[0_1px_2px_rgba(15,23,42,0.02)]"><span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[9px] bg-slate-100 text-slate-500"><Icon name={area.icon} size={17} /></span><span><span className="flex items-center gap-2 text-sm font-semibold text-slate-800">{area.label}<span className="rounded bg-slate-100 px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-wide text-slate-400">Soon</span></span><span className="mt-1.5 block text-xs leading-5 text-slate-500">{area.detail}</span></span></button>)}</div></section>
        <section className="mt-8 rounded-[14px] border border-dashed border-slate-300 bg-white px-5 py-6"><div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"><div><h2 className="text-[15px] font-semibold text-slate-900">Recent operations</h2><p className="mt-1 text-xs text-slate-500">Audit events will appear here once the protected admin activity feed is connected.</p></div><span className="inline-flex w-fit items-center rounded-full bg-slate-100 px-2.5 py-1 text-[11px] font-medium text-slate-500">No activity feed</span></div></section>
      </div>
    </main>
  </div></div>;
}
