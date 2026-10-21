"use client";
import { Button } from '@/components/ui/Button';
export default function ErrorPage({ reset }: { reset: () => void }) {
  return <div className="feed-container route-loading"><section className="feed-state" role="alert"><h1>Something didn&apos;t load.</h1><p>Please try again to return to your events.</p><Button onClick={reset}>Try again</Button></section></div>;
}
