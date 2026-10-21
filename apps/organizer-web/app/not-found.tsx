import Link from 'next/link';
export default function NotFound() { return <section className="desk-page desk-empty"><h1>This event isn’t available.</h1><p>Check the link and make sure you’re signed in with the account that hosts this event.</p><Link className="app-button app-button-primary" href="/">My hosted events</Link></section>; }
