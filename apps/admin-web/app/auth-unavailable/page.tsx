import type { Metadata } from 'next';

export const metadata: Metadata = { title: 'Sign-in unavailable | Hackways Admin' };

export default function AuthUnavailablePage() {
  return <main className="admin-login">
    <a href="https://hackways.in" className="wordmark">hackways.</a>
    <h1>Sign-in unavailable</h1>
    <p>We could not check your access. Please try again.</p>
    <a href="/login" className="app-button app-button-outline">Try again</a>
  </main>;
}
