"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import { Skeleton } from "@/components/ui/Skeleton";

interface HealthSnapshot {
  status: string;
  database?: string;
  version?: string;
  timestamp?: string;
  uptime?: string;
  environment?: string;
}

interface ServiceState {
  data: HealthSnapshot | null;
  error: string | null;
}

type HealthState = { core: ServiceState; global: ServiceState };

const unavailable: ServiceState = { data: null, error: "Not checked" };

async function readHealth(path: string): Promise<HealthSnapshot> {
  const response = await fetch(path, { cache: "no-store" });
  let body: HealthSnapshot;
  try {
    body = (await response.json()) as HealthSnapshot;
  } catch {
    throw new Error(`Health endpoint returned ${response.status} without a health payload.`);
  }
  if (!response.ok && !body.status) {
    throw new Error(`Health endpoint returned HTTP ${response.status}.`);
  }
  return body;
}

function HealthCard({ name, service }: { name: string; service: ServiceState }) {
  const healthy = service.data?.status === "ok" && service.data.database !== "unhealthy";
  return (
    <article className="rounded-xl border border-[#E4E7F0] bg-white p-6">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-[#0E121E]">{name}</h2>
        {service.data ? (
          <span className={`text-sm font-semibold ${healthy ? "text-emerald-800" : "text-rose-800"}`}>
            {healthy ? "Healthy" : service.data.status}
          </span>
        ) : (
          <span className="text-sm font-semibold text-rose-800">Unavailable</span>
        )}
      </div>
      {service.data ? (
        <dl className="mt-5 grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
          <dt className="text-[#757F9A]">Database</dt>
          <dd className="text-right font-medium text-[#0E121E]">{service.data.database ?? "Not reported"}</dd>
          <dt className="text-[#757F9A]">Uptime</dt>
          <dd className="text-right font-medium tabular-nums text-[#0E121E]">{service.data.uptime ?? "Not reported"}</dd>
          <dt className="text-[#757F9A]">Version</dt>
          <dd className="text-right font-mono text-xs text-[#475069]">{service.data.version ?? "Not reported"}</dd>
          <dt className="text-[#757F9A]">Environment</dt>
          <dd className="text-right text-[#475069]">{service.data.environment ?? "Not reported"}</dd>
        </dl>
      ) : (
        <p role="alert" className="mt-4 text-sm text-rose-800">{service.error}</p>
      )}
      {service.data?.timestamp && (
        <p className="mt-5 border-t border-[#E4E7F0] pt-3 text-xs text-[#757F9A]">
          Last checked {new Date(service.data.timestamp).toLocaleString()}
        </p>
      )}
    </article>
  );
}

export default function AdminPortalPage() {
  const [health, setHealth] = useState<HealthState>({ core: unavailable, global: unavailable });
  const [isLoading, setIsLoading] = useState(true);

  const loadHealth = useCallback(async () => {
    setIsLoading(true);
    const [coreResult, globalResult] = await Promise.allSettled([
      readHealth("/health/core"),
      readHealth("/health/global"),
    ]);
    setHealth({
      core: coreResult.status === "fulfilled"
        ? { data: coreResult.value, error: null }
        : { data: null, error: coreResult.reason instanceof Error ? coreResult.reason.message : "Health check failed." },
      global: globalResult.status === "fulfilled"
        ? { data: globalResult.value, error: null }
        : { data: null, error: globalResult.reason instanceof Error ? globalResult.reason.message : "Health check failed." },
    });
    setIsLoading(false);
  }, []);

  useEffect(() => {
    void loadHealth();
  }, [loadHealth]);

  return (
    <main className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
      <header className="flex flex-col justify-between gap-4 border-b border-[#E4E7F0] pb-7 sm:flex-row sm:items-end">
        <div>
          <p className="text-sm font-semibold text-[#34349C]">Platform operations</p>
          <h1 className="mt-1 text-3xl font-bold tracking-tight text-[#0E121E]">Service health</h1>
          <p className="mt-2 max-w-2xl text-sm text-[#475069]">
            Readiness and database status from each backend health endpoint.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={() => void loadHealth()} disabled={isLoading}>
          {isLoading ? "Checking…" : "Refresh health"}
        </Button>
      </header>

      <section aria-label="Backend health" className="grid gap-5 py-8 md:grid-cols-2">
        {isLoading && !health.core.data && !health.global.data ? (
          <>
            <div className="rounded-xl border border-[#E4E7F0] bg-white p-6"><Skeleton className="h-6 w-40" /><Skeleton className="mt-5 h-20 w-full" /></div>
            <div className="rounded-xl border border-[#E4E7F0] bg-white p-6"><Skeleton className="h-6 w-40" /><Skeleton className="mt-5 h-20 w-full" /></div>
          </>
        ) : (
          <>
            <HealthCard name="Core backend" service={health.core} />
            <HealthCard name="Global backend" service={health.global} />
          </>
        )}
      </section>
    </main>
  );
}
