export const metadata = { title: 'Access restricted', robots: { index: false, follow: false } };

export default function NotAllowedPage() {
  return <div className="feed-container route-loading"><section className="feed-state">
    <h1>You are not allowed</h1>
    <p>Your account or browser needs administrator review.</p>
    <a className="app-button app-button-outline" href="/help">Get help</a>
  </section></div>;
}
