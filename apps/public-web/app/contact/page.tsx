import { pageSocialMetadata } from '@/lib/social-metadata';
import type { Metadata } from 'next';
import Link from 'next/link';
import { Article, ArticleSection as Section } from '@/components/site/Article';
import { site, supportEmail } from '@/lib/site';
export const metadata: Metadata = { ...pageSocialMetadata('contact'), title: 'Contact Hackways', description: 'Contact Hackways support for account, platform, privacy and safety concerns, or find the organizer for event-specific questions.', alternates: { canonical: '/contact' } };
export default function ContactPage() {
  return <Article variant="contact" title="Contact Hackways" intro="Get your question to the right place. Contact the organizer about their event, or contact Hackways for help with your account, the platform, privacy or misuse.">
    <Section id="support" title="Contact platform support">
      <p>Email <a className="email-link" href={'mailto:' + supportEmail}>{supportEmail}</a> for an account problem, a registration result you cannot confirm, a broken feature or a concern about how Hackways is being used.</p>
      <p>Start your subject with the main issue, such as Account access, Registration help, Privacy request or Report an event. Include the event link when your question relates to a gathering.</p>
      <div className="section-actions"><a className="button button-primary" href={'mailto:' + supportEmail}>Email support</a><Link className="text-link" href="/help">Read Help & FAQs</Link></div>
    </Section>
    <Section id="event" title="Ask the organizer about their event">
      <p>The organizer is the right contact for the programme, venue, accessibility, age or eligibility requirements, travel, admission rules and online joining arrangements. Open the event page and look for its contact page or the information in its description and FAQs.</p>
      <p>If you need to change an eligible registration, first open its details in My events. Available actions can include cancellation, a note update or transfer. If the action you need is not offered, explain the request to the organizer.</p>
      <p>A private event link does not guarantee restricted admission. Confirm any invitation or entry requirement with the organizer before making plans.</p>
      <a className="text-link" href={site.app + '/my-events'}>Open My events</a>
    </Section>
    <Section id="include" title="Help us understand the problem">
      <ul><li>The event name and public event link, if relevant.</li><li>A short description of what you tried and what happened instead.</li><li>The approximate date and time, including your time zone.</li><li>The error message, and your browser or device type if it is a display or access issue.</li><li>Your own registration reference when it is needed to identify the booking. Send it privately to support rather than posting it publicly.</li></ul>
      <p>A cropped screenshot can be useful. Hide other guests, personal contact details and anything unrelated to the problem.</p>
      <div className="article-note"><p>Never send a password, verification code, session cookie, complete ticket QR code or payment card details. Support does not need those credentials to understand a report.</p></div>
    </Section>
    <Section id="privacy" title="Ask about your information">
      <p>Use the support email for questions about the information associated with your account or registration, or to request access, correction or deletion. Describe the records and the action you are asking for. Where possible, write from the email address used for the account or booking.</p>
      <p>Identity or ownership may need to be checked before private information can be disclosed or a record changed. Sending a request does not itself delete data. Read the <Link href="/privacy">privacy notice</Link> for the current scope of information and retention.</p>
      <p>For a copy of registration information held separately by an event organizer, contact that organizer as well.</p>
    </Section>
    <Section id="report" title="Report misuse or a security concern">
      <p>Send the event or page link and explain the concern: impersonation, misleading information, harassment, exposure of private information, spam or a possible security problem. Include only the evidence needed to understand it and avoid redistributing harmful or private material.</p>
      <p>For a technical vulnerability, describe a minimal reproduction using your own account and records. Do not access another account, download guest data, disrupt a live event or continue exploiting the issue.</p>
      <p>Hackways support is not an emergency service. If someone is in immediate danger, contact the appropriate local emergency service or venue security.</p>
    </Section>
    <Section id="expect" title="Keep the conversation in one place">
      <p>Keep a copy of your message and add any follow-up information to the same email thread. A published support response time is not currently specified, so do not rely on an immediate reply for time-sensitive event arrangements.</p>
      <p>An email link opens your mail application; Hackways does not receive a message until you send it. For an event happening soon, contact the organizer directly using the details on its page.</p>
    </Section>
  </Article>;
}
