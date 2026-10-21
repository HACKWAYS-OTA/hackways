import type { Metadata } from 'next';
import Link from 'next/link';
import { Article, ArticleSection as Section } from '@/components/site/Article';
import { site } from '@/lib/site';
export const metadata: Metadata = { title: 'For organizers', description: 'A practical guide to creating, publishing and managing free events, applications, waiting lists and guest check-in on Hackways.', alternates: { canonical: '/organizers' } };
export default function OrganizersPage() {
  return <Article variant="organizers" title="Organize an event" intro="From your first event page to your guest list on the day. Use this guide to prepare a clear invitation and understand the tools available for free events.">
    <Section id="start" title="Start with your account and host profile">
      <p>Sign in to the main Hackways app and choose Create event. Add your host profile if you have not set one up. Use a name guests will recognize and an email address you control.</p>
      <p>You can host a personal event or associate it with a Space you own. Event creation happens in the main app; the organizer workspace is where you manage events you have already created.</p>
      <div className="section-actions"><a className="button button-primary" href={site.app + '/create'}>Create an event</a><a className="text-link" href={site.organizer + '/login'}>Sign in to manage events</a></div>
    </Section>
    <Section id="details" title="Make the event easy to understand">
      <ul><li>Give it a specific name and a useful description: audience, programme, preparation and participation requirements.</li><li>Set the start, end and event time zone. Check any field marked UTC carefully; the public page presents the schedule in the event time zone.</li><li>Choose in person, online or hybrid. In-person and hybrid gatherings need a venue; online and hybrid gatherings need a secure joining link.</li><li>Add a street address where it helps people arrive. Include accessibility, travel or arrival information in the description or FAQs.</li><li>Add a contact page, FAQs and rewards information when relevant. State eligibility and conditions clearly.</li></ul>
      <p>You can personalize the cover, colour, style, font and light or dark theme. Use images you have permission to publish. A cover upload is complete only after the application confirms it; keep the form open and follow any recovery message if saving fails.</p>
    </Section>
    <Section id="registration" title="Choose how people join">
      <dl className="info-definitions"><div><dt>Direct registration</dt><dd>A place can be confirmed when the guest submits a valid form and capacity is available. You can allow guest registration without an account.</dd></div><div><dt>Approval</dt><dd>Guests sign in to apply. Their application waits for your review; submitting it does not reserve a place. Review applications in your guest list.</dd></div><div><dt>Capacity</dt><dd>Set a limit or leave it empty for no limit. Places already reserved cannot simply be removed by lowering the capacity.</dd></div><div><dt>Waiting list</dt><dd>For direct registration, you can enable a waiting list when the event fills up. Guests need an account to join. An offer is time limited and must be accepted before it expires.</dd></div><div><dt>Registration window</dt><dd>Optionally set when registration opens and closes. Joining also depends on the event&apos;s publication state, start time and available places.</dd></div></dl>
      <p>Ask only for information you need. Registration questions support short and long answers, a single choice, checkbox, number, website and phone number fields. Mark genuinely necessary questions as required. Changing a question does not rewrite a guest&apos;s historical answer.</p>
    </Section>
    <Section id="publish" title="Check visibility before you share">
      <p>Save a draft while you are preparing, then publish when the details are ready. Public events can appear in Discover and their Space&apos;s public event list. Private events are shared by link and are left out of public discovery.</p>
      <p>A private link is not an invitation-only access list. Anyone who has the link may be able to view the page and register under its settings. Avoid placing confidential information on an event page.</p>
      <p>Open View event after saving and check the guest-facing title, schedule, venue and description. Share that event link with attendees.</p>
    </Section>
    <Section id="manage" title="Work from the guest list">
      <p>Open an event in the organizer workspace to search guests, filter by registration or arrival status, and inspect individual details. A guest record includes their registration information and the answers they submitted.</p>
      <p>Review pending applications and use the available status actions. Where status editing is offered, choose the new status, decide whether to notify the guest, and optionally add a message. Confirmed, checked-in or otherwise protected records may have different available actions.</p>
      <p>Notifications are queued after a successful operation. A queued message is not proof that an email reached the recipient. Use the saved guest status as the source of truth and address any delivery problem separately.</p>
      <p>Guest details and answers are private. Use them to administer the event and keep any copies you make secure.</p>
    </Section>
    <Section id="event-day" title="Prepare for check-in">
      <p>Before the event, confirm the latest venue and joining arrangements with your team. Use the check-in tools in the event workspace to validate eligible tickets and record arrival.</p>
      <p>A ticket can be checked in only once. If a manual check-in is needed, the organizer must provide the required reason. A pending application or waiting-list entry is not a confirmed ticket.</p>
      <p>Plan a fallback for guests who cannot find their registration and for venue connectivity problems. Offline scanning is not part of the current check-in workflow.</p>
    </Section>
    <Section id="changes" title="Keep guests informed when plans change">
      <p>Use the event editor to save supported changes. For schedule or location changes to a published event, use the notification option when appropriate. Review the save result before leaving the page; a failure keeps recoverable entries available for retry.</p>
      <p>Pausing registration and cancelling an event have different consequences. A pause stops new registration without cancelling existing bookings. Cancelling an event closes registration, invalidates tickets and queues attendee notifications; review the confirmation carefully.</p>
      <p>If another edit was saved in the meantime, follow the conflict message and reload the current version before trying again. Do not assume an unsaved preview is what guests see.</p>
      <p>For an account or platform problem, visit <Link href="/contact">Contact</Link>. For registration questions, the <Link href="/help/hosting">hosting FAQs</Link> provide more guidance.</p>
    </Section>
  </Article>;
}
