'use client';
export default function AdminError({ reset }: { reset: () => void }) { return <section className="admin-result-state" role="alert"><h1>Admin data is unavailable</h1><p>We could not load this page. Try again in a moment.</p><button type="button" className="app-button app-button-outline" onClick={reset}>Try again</button></section>; }
