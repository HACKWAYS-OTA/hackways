import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { RolloutEditor } from '@/components/admin/RolloutEditor';
import { adminServer } from '@/lib/admin-server';
import type { RolloutData } from '@/lib/rollouts';
export const metadata = { title: 'Feature rollouts & versions | Hackways Admin' };
export default async function Rollouts({ searchParams }: { searchParams: { offset?: string } }) {
  const offset = Math.max(0, Number(searchParams.offset) || 0);
  const result = await adminServer<RolloutData>('global', `/rollouts?offset=${offset}`);
  return <AdminShell active="/rollouts"><div className="admin-page-heading"><h1>Feature rollouts & versions</h1></div><section className="admin-build-provenance" aria-label="Running build provenance"><span>Evaluator: {result.cache.fresh ? `Version #${result.cache.version}, refreshed ${result.cache.age_seconds}s ago` : 'Safe baseline active; snapshot unavailable or expired'}</span><span>Global engine source: <code>{result.source_commit}</code></span><span>Admin web source: <code>{process.env.SOURCE_COMMIT || 'Unreported local build'}</code></span><p>Configuration versions control features included in the running application. Application image releases remain in the approved CI deployment pipeline.</p></section><RolloutEditor key={result.data.revision} initial={result} /><nav className="admin-pagination" aria-label="Configuration history pages">{offset > 0 && <Link className="app-button app-button-outline" href={`/rollouts?offset=${Math.max(0, offset - 20)}`}>Previous</Link>}{result.has_next && <Link className="app-button app-button-outline" href={`/rollouts?offset=${offset + 20}`}>Next</Link>}</nav></AdminShell>;
}
