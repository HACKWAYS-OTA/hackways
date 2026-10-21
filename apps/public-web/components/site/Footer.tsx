import Link from "next/link";
import { site, supportEmail } from "@/lib/site";
import { Brand } from "./Brand";
import './footer.css';

export function Footer() {
  return (
    <footer className="site-footer">
      <div className="site-container">
        <div className="footer-main">
          <div className="footer-brand"><Brand /><p>A place to find your next event.<br />And the people who make it worthwhile.</p></div>
          <nav aria-label="Explore" className="footer-column"><h2>Explore</h2><a href={`${site.app}/discover`}>Find an event</a><Link href="/organizers">For organizers</Link><Link href="/about">About Hackways</Link></nav>
          <nav aria-label="Resources" className="footer-column"><h2>Resources</h2><Link href="/help">Help & FAQs</Link><Link href="/contact">Contact</Link><a href={`${site.organizer}/login`}>Organizer sign in</a><a href={`mailto:${supportEmail}`}>{supportEmail}</a></nav>
          <nav aria-label="Legal" className="footer-column"><h2>The details</h2><Link href="/privacy">Privacy</Link><Link href="/terms">Terms of use</Link><Link href="/cookies">Cookies</Link><Link href="/acceptable-use">Acceptable use</Link></nav>
        </div>
        <div className="footer-bottom"><p>© {new Date().getFullYear()} Hackways</p><p>Good things happen together.</p></div>
      </div>
    </footer>
  );
}
