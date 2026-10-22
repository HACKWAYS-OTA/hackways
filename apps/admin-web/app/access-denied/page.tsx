import type { Metadata } from 'next';

export const metadata: Metadata = { title: 'Access denied | Hackways Admin' };

export default function AccessDeniedPage() {
  return <main className="admin-login">
    <a href="https://hackways.in" className="wordmark">hackways.</a>
    <h1>Access denied</h1>
    <p>This account does not have access to Hackways Admin.</p>
    <a href="/login?switch=1" className="app-button app-button-outline">Use another account</a>
  </main>;
}
