import { pageSocialMetadata } from '@/lib/social-metadata';
import type { Metadata } from 'next';
import Link from 'next/link';
import { CookiePreferencesButton } from '@hackways/ui/analytics';
import { Article, ArticleSection as Section } from '@/components/site/Article';
import { supportEmail } from '@/lib/site';
export const metadata: Metadata = { ...pageSocialMetadata('cookies'), title: 'Cookies', description: 'The essential session cookies used by Hackways, how long sessions last, and how to manage cookies and sign-in in your browser.', alternates: { canonical: '/cookies' } };
export default function CookiesPage() {
  return <Article title="Cookies" intro="What your browser stores when you use Hackways, why account sessions need cookies, and how you can control them." updated="23 October 2026">
    <Section id="preferences" title="Manage cookies"><p>Choose whether this browser allows optional analytics. Essential cookies keep sign-in working.</p><CookiePreferencesButton label="Manage cookies" /></Section>
    <Section id="purpose" title="Cookies help a service recognize a browser">
      <p>A cookie is a small value stored by a browser for a website. It can be sent back with later requests so a service can recognize a session. Cookies are different from the account and registration records stored by the service.</p>
      <p>The account applications use an essential cookie for sign-in. The public informational pages can be read without an account. Optional first-party analytics runs only when you choose Allow all.</p>
    </Section>
    <Section id="public" title="Reading the public website">
      <p>You do not need to sign in to read these pages. Optional analytics cookies are off until you allow them. No third-party advertising tracker is included. Fonts used by these pages are served from the website.</p>
      <p>Following Find an event or Host an event takes you to the relevant Hackways app. That app may use a sign-in cookie when you create an account or log in. Infrastructure may still process network requests as described in the <Link href="/privacy">privacy notice</Link>; the absence of tracking cookies does not mean no information is processed.</p>
    </Section>
    <Section id="session" title="The essential account session cookie">
      <dl className="info-definitions"><div><dt>Name</dt><dd><code>__Host-hackways_session</code></dd></div><div><dt>Purpose</dt><dd>Associates browser requests with a signed-in session so you can access your account, connected registrations or event-management tools.</dd></div><div><dt>Duration</dt><dd>A normal account session is issued for up to ninety days. It can end earlier when you sign out, revoke a session or the service no longer accepts it.</dd></div><div><dt>Scope</dt><dd>Production sessions are scoped to the application host. The attendee app and organizer workspace can therefore ask you to sign in separately.</dd></div><div><dt>Protection</dt><dd>The cookie uses Secure and HTTP-only attributes and a SameSite policy. Ordinary page scripts cannot read the HTTP-only cookie value. These settings do not replace protecting your device and account.</dd></div></dl>
      <p>The cookie contains a session reference rather than your readable password. Do not copy or share it: someone who obtains a usable session credential may be able to act as that signed-in session.</p>
      <p>Signing out clears the cookie for that application and revokes the presented session. Account session controls can also revoke your sessions. Signing out on one production host does not automatically clear a cookie on every other host.</p>
    </Section>
    <Section id="analytics" title="Optional first-party analytics cookies">
      <dl className="info-definitions"><div><dt>Choice</dt><dd><code>hackways_analytics_consent</code> remembers your yes or no choice for up to one year.</dd></div><div><dt>Visitor</dt><dd><code>hackways_analytics_visitor</code> is a random browser identifier, refreshed for up to ninety days while analytics is allowed.</dd></div><div><dt>Visit</dt><dd><code>hackways_analytics_session</code> groups activity for up to thirty minutes of inactivity.</dd></div></dl>
      <p>Collection covers approved public page paths, viewport category, referring hostname and registration progress. It excludes form answers, account identifiers, query strings and full referring URLs. Browser identifiers are hashed by the service before storage. Raw analytics events are removed after ninety days by a bounded cleanup job.</p>
      <p>Changing this choice affects this browser. Previously collected events are not automatically deleted when you decline future collection. Browser restrictions or blocked requests may prevent collection, so analytics is not a complete visitor count.</p>
    </Section>
    <Section id="controls" title="Manage cookies in your browser">
      <ol><li>Use the sign-out option when you finish on a shared device.</li><li>Open your browser settings for cookies or site data to inspect, remove or block cookies for a site.</li><li>If you remove an account session cookie, expect to sign in again before using account-only features.</li><li>If sign-in is not working, check whether cookies are blocked for the app you are opening and whether the session has expired.</li></ol>
      <p>Blocking essential cookies can prevent sign-in and features that require an account. Clearing browser data does not delete event registrations or account records held by Hackways.</p>
      <p>For current browser instructions, use the official <a href="https://support.google.com/chrome/answer/95647">Chrome cookie settings guide</a> or <a href="https://support.mozilla.org/en-US/kb/clear-cookies-and-site-data-firefox">Firefox cookie and site-data guide</a>. Other browsers provide equivalent site-data controls in their settings.</p>
    </Section>
    <Section id="external" title="Links to organizer and third-party services">
      <p>An event may link to a video meeting, venue, organizer website or another external service. Once you open that destination, its own cookie and privacy practices apply. Review those notices before sharing more information or accepting optional tracking there.</p>
      <p>Use Manage cookies on this page to allow optional analytics or switch to Essential only. Declining removes the analytics browser identifiers and stops further collection in that browser; essential sign-in cookies remain available.</p>
    </Section>
    <Section id="questions" title="Ask about cookies or account information">
      <p>Email <a href={'mailto:' + supportEmail}>{supportEmail}</a> for questions about these cookies. Read <Link href="/privacy">Privacy</Link> for account and registration information, and <Link href="/contact">Contact</Link> for guidance on making a records request.</p>
    </Section>
  </Article>;
}
