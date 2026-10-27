import { pageSocialMetadata } from '@/lib/social-metadata';
import type { Metadata } from 'next';
import Link from 'next/link';
import { Article, ArticleSection as Section } from '@/components/site/Article';
import { supportEmail } from '@/lib/site';
export const metadata: Metadata = { ...pageSocialMetadata('privacy'), title: 'Privacy', description: 'How Hackways uses account, event, registration, ticket and session information, who can access it, and how to raise a privacy request.', alternates: { canonical: '/privacy' } };
export default function PrivacyPage() {
  return <Article title="Privacy" intro="Understand what information is used when you browse, register, host an event or sign in to Hackways, and how to ask about your records." updated="23 October 2026">
    <Section id="scope" title="What this notice covers">
      <p>This notice describes the current Hackways public website, attendee app and organizer workspace. It covers information handled through those services. An event organizer or an external website may handle information separately under its own arrangements.</p>
      <p>You can read public event information without creating an account. Registration and hosting require additional information, and some events require you to sign in. The relevant form shows which details are required.</p>
    </Section>
    <Section id="information" title="Information used by the service">
      <dl className="info-definitions"><div><dt>Account details</dt><dd>Your display name, email address, password hash, email-verification status and account records. A password is stored as a one-way hash rather than as readable text.</dd></div><div><dt>Registration</dt><dd>Your name or account identity, email address, optional note, answers to the event registration questions, registration reference, status and associated timestamps. Required fields depend on the event form.</dd></div><div><dt>Tickets and arrival</dt><dd>Ticket references, issuance or revocation, transfer and offer records where relevant, and a check-in record when your ticket is admitted.</dd></div><div><dt>Organizer content</dt><dd>Host and Space profiles, community membership, event titles and descriptions, schedules, locations, joining and contact links, covers, FAQs, rewards and publication settings.</dd></div><div><dt>Session and service records</dt><dd>Session identifiers and expiry, browser user-agent and request network information used for sessions, security and operation. Event administration can also create action and revision history.</dd></div><div><dt>Communication</dt><dd>Notification recipient details, message content where a notification includes a custom message, delivery-processing records and any information you choose to send in a support email.</dd></div></dl>
      <p>The event editor may read the time zone provided by your browser to suggest a default. Avoid putting passwords, financial credentials or unnecessary sensitive information in event descriptions, answers or notes.</p>
    </Section>
    <Section id="use" title="Why information is used">
      <ul><li>Recognize your account, authenticate a session and connect registrations to the correct person.</li><li>Publish event and Space information and present the schedule and joining arrangements.</li><li>Process registrations, enforce capacity, review applications and manage waiting-list offers.</li><li>Provide eligible tickets, transfers, cancellation and check-in actions.</li><li>Queue and process transactional messages such as verification, registration or event updates.</li><li>Investigate errors, protect access to private records and keep administration history.</li></ul>
      <p>Submitting a form does not mean every requested action has succeeded. The application displays the saved result or an error. An email being queued is separate from successful delivery.</p>
    </Section>
    <Section id="visibility" title="Public information and private guest records">
      <p>Published event content is intended for people who can open the event page. Public events may also appear in discovery and public Space feeds. Private events are shared by link and are not an access-controlled invitation list; do not publish confidential material there.</p>
      <p>The event owner can view guest identities, registration status and submitted answers for their event. Those guest records are not a public attendee directory. Joining a Space does not grant access to its guest lists or management tools.</p>
      <p>Authorized platform operations may need access to records to investigate a support, security or service issue. Information may also need to be provided in response to a valid legal requirement. Sharing an event link or making a support request does not grant unrelated people access to private registrations.</p>
      <p>If an organizer copies information into a separate system, ask that organizer about its use and retention there. Hackways cannot describe an independent recipient as operating under this notice.</p>
    </Section>
    <Section id="providers" title="Infrastructure, email and external sites">
      <p>Hosting, database, media storage and email services process information needed to operate Hackways. Event cover media uses Azure Blob Storage, and transactional email uses Azure Communication Services. Configured infrastructure and service delivery may involve processing outside your immediate location; this notice does not promise a particular data-residency region.</p>
      <p>Optional first-party analytics collects approved public and managed-workflow page names (with private resource identifiers removed), viewport category, referring hostname and registration progress only after you allow analytics. Browser identifiers are hashed before storage; these events do not include account identifiers, IP addresses, form answers, query strings or full referring URLs. Infrastructure and security logs remain separate. Signed-in apps also use essential session cookies, explained in the <Link href="/cookies">cookie notice</Link>.</p>
      <p>Event pages may link to venues, video meetings, organizer contact pages or other external services. Opening those links is subject to the destination service and its privacy practices. Check the destination before submitting additional personal or payment information.</p>
    </Section>
    <Section id="analytics-choices" title="Analytics and promotional email choices">
      <p>Use Cookie preferences in the footer to allow optional analytics or choose Essential only. Analytics browser identifiers last up to ninety days, visit identifiers up to thirty minutes of inactivity, and your choice up to one year. Raw analytics events are removed after ninety days by a bounded cleanup job. Visitor charts cover only collected activity; saved account and registration totals come from business records.</p>
      <p>Promotional email starts enabled for new accounts. The email signup form lets you turn it off before creating your account. Existing unsubscribes are preserved. You can change promotional email in your account or use the unsubscribe link included in promotional messages. Essential account and event messages are handled separately.</p>
    </Section>
    <Section id="storage" title="Security and retention">
      <p>Access to account and event-management actions is checked by the service. Passwords are hashed, and session cookies use HTTP-only and secure attributes. These measures reduce risk but are not a promise that every device, account or transmission is free of security risk.</p>
      <p>Use a unique password, protect your email account and sign out of shared devices. Keep tickets, verification codes and session credentials private. If you suspect unauthorized access, use the account session controls when available and contact support.</p>
      <p>Account, registration and administration records can remain stored after an event ends. An event end date, a cancellation or the expiry of a browser cookie is not a deletion deadline for those records. A fixed retention schedule is not specified in this notice.</p>
      <p>Deleting cookies does not delete your account or booking records. Requests about retention or deletion should identify the records concerned; legal, security, operational or backup considerations may need to be assessed before a request is resolved.</p>
    </Section>
    <Section id="choices" title="Your choices and information requests">
      <p>You can choose what to include in optional fields, manage available account sessions, and use the registration actions shown for your booking. Required event questions must be answered to submit that form. Contact the organizer if a question seems unnecessary or unclear.</p>
      <p>For questions or requests concerning access, correction or deletion of information held through Hackways, email <a href={'mailto:' + supportEmail}>{supportEmail}</a>. Describe the records and the outcome you are requesting, preferably from the address associated with them.</p>
      <p>We may need to check ownership before disclosing private information or changing a record. An email request is not an automatic deletion or export operation. Do not include a password, session token, verification code or complete ticket QR code.</p>
      <p>Applicable privacy rights and the handling of a request depend on the circumstances and relevant law. This notice does not replace those rights or claim that a self-service account deletion or export feature exists.</p>
    </Section>
    <Section id="updates" title="Questions and changes to this notice">
      <p>Send privacy concerns to <a href={'mailto:' + supportEmail}>{supportEmail}</a>. The <Link href="/contact">contact guide</Link> explains what to include. Contact the event organizer separately when your request concerns information it holds outside Hackways.</p>
      <p>This page may be updated as features or operational arrangements change. The date above identifies the current text; revisit it when deciding how to use a new feature.</p>
    </Section>
  </Article>;
}
