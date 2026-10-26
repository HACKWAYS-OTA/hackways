import Link from 'next/link';
export default function NotFound() { return <section className="desk-empty"><h1>This City Hub isn’t available.</h1><Link className="app-button app-button-outline" href="/cities">Explore cities</Link></section>; }
