'use client';
export default function ProgramError({ reset }: { reset: () => void }) { return <section className="program-unavailable" role="alert"><h1>Programs couldn’t load.</h1><p>We couldn’t reach the program service. Please try again.</p><button className="app-button app-button-primary" type="button" onClick={reset}>Try again</button></section>; }
