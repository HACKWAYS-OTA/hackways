"use client";
export default function ErrorPage({ reset }: { reset: () => void }) { return <div className="admin-result-state" role="alert"><p>Campaign data could not be loaded.</p><button type="button" className="app-button app-button-outline" onClick={reset}>Try again</button></div>; }
