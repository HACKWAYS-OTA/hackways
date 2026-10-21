import Link from 'next/link';
export default function NotFound() {
  return <div className="feed-container route-loading"><section className="feed-state"><h1>This page isn&apos;t here.</h1><p>Head back to discovery to find upcoming events.</p><Link className="app-button app-button-primary" href="/discover">Discover events</Link></section></div>;
}
