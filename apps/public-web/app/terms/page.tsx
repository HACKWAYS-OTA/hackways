import { pageSocialMetadata } from '@/lib/social-metadata';
import type { Metadata } from 'next';
import Link from 'next/link';
import { Article, ArticleSection as Section } from '@/components/site/Article';
import { supportEmail } from '@/lib/site';
export const metadata: Metadata = { ...pageSocialMetadata('terms'), title: 'Terms of use', description: 'Responsibilities for accounts, event content, registration, guest information and participation when using Hackways.', alternates: { canonical: '/terms' } };
export default function TermsPage() {
  return <Article title="Terms of use" intro="The responsibilities that come with discovering, publishing and registering for events on Hackways. Read the event-specific details as well as these platform terms." updated="20 October 2026">
    <Section id="service" title="Understand the service you are using">
      <p>Hackways provides event discovery, publishing, registration and organizer tools. Organizers create and run their events and provide the programme, venue or online meeting, participation requirements and any event-specific promises.</p>
      <p>The current registration tools cover free events. A published listing does not make Hackways the event organizer or guarantee a particular experience, professional opportunity, reward or result.</p>
      <p>Use the service in accordance with these terms, the <Link href="/acceptable-use">acceptable use guidelines</Link> and applicable law. Read the <Link href="/privacy">privacy notice</Link> to understand how account and registration information is handled.</p>
    </Section>
    <Section id="account" title="Keep your account accurate and secure">
      <ul><li>Use an email address you control and provide accurate account information.</li><li>Keep your password and verification codes private. Do not share sessions or sign in as someone else.</li><li>Take responsibility for actions you authorize through your account, including event publication and guest administration.</li><li>Use the available session controls and contact support if you suspect unauthorized access.</li></ul>
      <p>Some event actions require an account or a verified email address. A registration made with another address does not become yours just because you know its reference or have a copy of its ticket.</p>
    </Section>
    <Section id="organizers" title="Publish clear and truthful event information">
      <p>As an organizer, you are responsible for the accuracy of your host identity, event listing and participation requirements, and for arranging and delivering the event. Obtain the permissions needed for your venue, media, programme and any third-party material you use.</p>
      <p>Describe dates, time zones, locations, joining instructions, eligibility, accessibility arrangements and rewards accurately. Explain material restrictions before guests register. Do not imply an endorsement, sponsorship or affiliation you do not have.</p>
      <p>Choose capacity, approval, waiting-list and visibility settings carefully. Public events may appear in discovery. Private links are not an invitation-only access system and may be forwarded by recipients.</p>
    </Section>
    <Section id="registration" title="Read the saved registration result">
      <p>A place is reserved only when the service confirms it. Browsing an event or starting a registration does not hold capacity. An application awaiting approval and a waiting-list entry are not confirmed places.</p>
      <p>Time-limited offers must be accepted before their displayed deadline. Ticket availability depends on registration status. Cancelled, expired, rejected, revoked or transferred entitlements do not provide admission, and a valid ticket can be checked in only once.</p>
      <p>Provide accurate details and answers. Do not create duplicate or fabricated registrations to hold places, bypass an approval decision or interfere with another attendee. Use a supported transfer action where offered rather than simply forwarding a ticket.</p>
      <p>For eligible bookings, cancellation and other actions appear in the registration details. Cancelling releases the place; registering again is subject to the event settings and availability at that time.</p>
    </Section>
    <Section id="changes" title="Check event changes before making plans">
      <p>Organizers may update supported details, pause registration or cancel an event. They are responsible for communicating material changes to guests and handling questions about the event arrangements.</p>
      <p>A pause and a cancellation have different effects. Cancellation closes the event and invalidates tickets. Review the current event page and your registration status before travelling or joining online.</p>
      <p>Transactional messages may be delayed or fail to arrive. A notification being queued is not a guarantee of delivery. For urgent programme, venue or admission questions, contact the organizer using the published event details.</p>
    </Section>
    <Section id="content" title="Respect content ownership and guest privacy">
      <p>Only upload or publish material you have the right to use. Publishing content makes it available through the visibility settings you choose and to the services needed to display it. Do not post private guest records, access credentials or confidential joining information in publicly accessible descriptions.</p>
      <p>Organizer access to registration information is for administering the event. Keep guest details and answers secure, use them responsibly, and do not sell, expose or repurpose them for unsolicited messages.</p>
      <p>If you copy data to another system or use an external service, you are responsible for the arrangements and obligations associated with that use. An event edit does not rewrite answers already submitted by a guest.</p>
    </Section>
    <Section id="conduct" title="Use the platform responsibly">
      <p>Do not impersonate people, harass others, publish deceptive events, distribute malicious content, bypass access or capacity controls, scrape private guest data or disrupt the service. The <Link href="/acceptable-use">acceptable use guidelines</Link> give further examples.</p>
      <p>Report a concern through <Link href="/contact">Contact</Link>, using the relevant page link and factual details. A report does not automatically establish a violation or guarantee a particular outcome. Do not expose another person to harm while collecting evidence.</p>
    </Section>
    <Section id="availability" title="Allow for service and third-party limitations">
      <p>Network connections, browsers, email delivery and external services can fail. Follow the actual success or error message after an action and check saved state before retrying an uncertain operation. A preview or unsaved form is not a published update.</p>
      <p>Events can link to external meeting, venue, contact or other websites. Those destinations have their own terms and practices. Verify the destination and organizer before sharing additional information or making an external payment.</p>
      <p>Nothing on this page is intended to remove rights or obligations that cannot be excluded under applicable law. Specific legal questions depend on the circumstances; these terms do not promise a result for an event or a dispute.</p>
    </Section>
    <Section id="questions" title="Questions about these terms">
      <p>Email <a href={'mailto:' + supportEmail}>{supportEmail}</a> with questions about the platform terms. Direct event-specific questions to the organizer. The <Link href="/help">help guide</Link> explains current features and the <Link href="/contact">contact guide</Link> explains how to report a problem.</p>
      <p>These terms may be updated as the platform changes. The date above identifies this version. Review the current terms and relevant event information when you use a new feature or join another event.</p>
    </Section>
  </Article>;
}
