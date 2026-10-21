'use client';
export default function ErrorPage({ reset }: { reset: () => void }) { return <section className="desk-page desk-empty"><h1>We couldn’t open this page.</h1><p>The event service may be unavailable. Your saved events are still kept on the server.</p><button className="app-button app-button-primary" onClick={reset}>Try again</button></section>; }
