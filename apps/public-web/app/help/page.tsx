import { pageSocialMetadata } from '@/lib/social-metadata';
import type { Metadata } from 'next';
import Link from 'next/link';
import { Article } from '@/components/site/Article';
import { FAQ, faqTopics } from '@/components/site/FAQ';
import { site, supportEmail } from '@/lib/site';
export const metadata: Metadata = { ...pageSocialMetadata('help'), title: 'Help & FAQs', description: 'Help with event discovery, RSVP status, waiting lists, tickets, cancellation, accounts and hosting on Hackways.', alternates: { canonical: '/help' } };
export default function HelpPage() {
  return <Article variant="help" title="How can we help?" intro="Find an answer, check your registration or get in touch. Start with the topic that matches what you need.">
    <p>For the status of your own registration, start with <a href={site.app + '/my-events'}>My events</a>. For a programme, venue or admission question, check the contact information on the event page.</p>
    <nav className="help-topics" aria-label="Help topics">{faqTopics.map(topic => <Link className="help-topic" key={topic.id} href={`/help/${topic.id}`}><h2>{topic.label}</h2><p>{topic.intro}</p></Link>)}</nav>
    <section className="popular-questions" aria-labelledby="popular-heading"><h2 id="popular-heading">Common questions</h2><FAQ /></section>
    <section className="article-note"><h2>Still need help?</h2><p>Email <a href={'mailto:' + supportEmail}>{supportEmail}</a> with the event link and a short description. The <Link href="/contact">contact guide</Link> explains what to include and how to keep sensitive details out of your request.</p></section>
  </Article>;
}
