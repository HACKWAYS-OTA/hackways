import Link from 'next/link';
export default function EventNotFound() { return <div className="feed-container route-loading"><section className="feed-state"><h1>This event isn&apos;t available.</h1><p>The link may have changed, or the event may no longer be published.</p><Link href="/discover" className="app-button app-button-primary">Discover events</Link></section></div>; }
