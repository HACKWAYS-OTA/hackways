import type { Metadata } from 'next';
import Link from 'next/link';
import { Article, ArticleSection as Section } from '@/components/site/Article';
import { site } from '@/lib/site';
export const metadata: Metadata = { title: 'About Hackways', description: 'Find technology events, understand how registration works, and meet the communities hosting on Hackways.', alternates: { canonical: '/about' } };
export default function AboutPage() {
  return <Article variant="about" title="About Hackways" intro="A place to discover technology events, bring people together, and keep the practical details of attending or hosting in one place.">
    <Section id="what-we-do" title="Events give people a reason to meet">
      <p>Hackways connects people with hackathons, meetups, conferences and other technology gatherings. You can explore published events, read the details, and register for something you want to be part of.</p>
      <p>For organizers, Hackways provides an event page and tools for managing registrations, reviewing applications and welcoming guests. The organizer provides the event itself: its programme, venue, joining arrangements and participation requirements.</p>
    </Section>
    <Section id="attending" title="Find something worth making time for">
      <p>Browse events by the topics that interest you, search event titles and descriptions, and choose between in-person, online and hybrid formats. An event page brings together the organizer, schedule, time zone, location or joining information, and the description.</p>
      <ul><li>Check the full event details and any participation requirements before registering.</li><li>Submit the registration form. Some events allow guest registration; others require an account or organizer approval.</li><li>Read the result carefully: a confirmed place, an application awaiting approval and a waiting-list entry are different outcomes.</li><li>Use My events when signed in to follow the registrations connected to your account and see the actions available for each one.</li></ul>
      <p>Opening a page or beginning a form does not reserve a place. Confirmation depends on the event&apos;s settings and availability when your request is processed.</p>
      <a className="text-link" href={site.app + '/discover'}>Explore events</a>
    </Section>
    <Section id="hosting" title="Give your gathering a clear home">
      <p>Create an event with its title, description, schedule and joining details. Add a cover and a theme, set a capacity if needed, and decide whether guests join directly or apply for approval.</p>
      <p>Publish publicly to appear in discovery, or use a private event link for a gathering you want to share directly. Manage your existing events in the organizer workspace, where the guest list, registration settings and check-in tools are available to the event owner.</p>
      <p>Good event information answers practical questions early: who it is for, what will happen, how to arrive or join, and whom to contact if a guest needs help.</p>
      <Link className="text-link" href="/organizers">Read the organizer guide</Link>
    </Section>
    <Section id="spaces" title="Stay connected through a Space">
      <p>A Space gives a community a profile and a place for its published events. You can browse Spaces and join the ones you are interested in. Organizers can associate their events with a Space they own.</p>
      <p>Space membership does not give someone permission to manage its events or see private guest records. A private event remains absent from public discovery and public Space event feeds.</p>
      <a className="text-link" href={site.app + '/spaces'}>Browse Spaces</a>
    </Section>
    <Section id="expectations" title="Know what you are signing up for">
      <p>Event listings are supplied by their organizers. Review the latest page before travelling or joining online, especially if the time or location may have changed. Follow the organizer&apos;s event rules and treat other participants with respect.</p>
      <p>The current registration tools cover free events. A listing is not a promise of a particular experience, outcome, prize or affiliation. Any rewards, eligibility rules or event-specific commitments should be described by the organizer.</p>
      <p>Account and registration information is used to operate these journeys. Our <Link href="/privacy">privacy notice</Link> explains what is collected and who can see it; the <Link href="/acceptable-use">acceptable use guidelines</Link> describe the conduct expected on the platform.</p>
    </Section>
    <Section id="next-step" title="Choose where to begin">
      <p>Discover a gathering, start planning your own, or get an answer before you take the next step.</p>
      <div className="section-actions"><a className="button button-primary" href={site.app + '/discover'}>Find an event</a><a className="button button-outline" href={site.app + '/create'}>Create an event</a><Link className="text-link" href="/help">Get help</Link></div>
    </Section>
  </Article>;
}
