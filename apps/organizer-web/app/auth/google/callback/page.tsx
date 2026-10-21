import { GoogleCallback } from '@/components/account/GoogleCallback';

export const metadata = { title: 'Google sign-in', robots: { index: false, follow: false } };

export default function GoogleCallbackPage({ searchParams }: { searchParams: { google_exchange?: string; google_error?: string; next?: string } }) {
  return <GoogleCallback code={searchParams.google_exchange} error={searchParams.google_error} next={searchParams.next} />;
}
