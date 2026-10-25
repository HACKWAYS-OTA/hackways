import { pageSocialMetadata } from '@/lib/social-metadata';
import type { Metadata } from 'next';
import Link from 'next/link';
import { HugeiconsIcon } from '@hugeicons/react';
import ArrowUpRight01Icon from '@hugeicons/core-free-icons/ArrowUpRight01Icon';
import ArrowDown01Icon from '@hugeicons/core-free-icons/ArrowDown01Icon';
import { SectionJumpButton } from '@hackways/ui/section-jump';
import { site } from '@/lib/site';
import { AmbassadorsHeroArt } from '@/components/ambassadors/AmbassadorsHeroArt';
import './ambassadors.css';

export const metadata: Metadata = { ...pageSocialMetadata('ambassadors'),
  title: 'Ambassadors',
  description: 'Bring your community together as a Hackways ambassador. Host your events on Hackways and receive recognition for your contributions.',
  alternates: { canonical: '/ambassadors' },
};

const joinURL = `${site.app}/programs/ambassadors/apply/profile`;

const responsibilities = [
  { title: 'Start with a shared interest.', description: 'Bring an idea your community would enjoy getting together around. A meetup, a workshop or a hackathon can be a place to start.' },
  { title: 'Host your event on Hackways.', description: 'Use Hackways for the events you host as an ambassador. Create the event page, share the invitation and manage your registrations in one place.' },
  { title: 'Bring the gathering to life.', description: 'Keep your event details clear, welcome your attendees and help people connect through the experience you organize.' },
];

const questions = [
  { question: 'Who can join the program?', answer: 'The program is open to students, professionals and community organizers. You do not need a particular college affiliation. Eligibility requirements are listed in the application.' },
  { question: 'What is my responsibility as an ambassador?', answer: 'Host your events using Hackways. You organize the gathering; Hackways gives you a place to publish the event, share its details and manage registrations.' },
  { question: 'What recognition is available?', answer: 'Your contributions are recognised with a letter of appreciation, with swag for the highest contributors.' },
  { question: 'How do I join?', answer: 'Sign in to Hackways, complete your profile and event proposal, then review and submit your application. You can save a draft and come back before submitting.' },
  { question: 'When will I hear back?', answer: 'We will send you a confirmation after you apply and share program updates by email. Results will be shared at the end of November.' },
];

function JoinLink({ label = 'Join the program', className = '' }: { label?: string; className?: string }) {
  return <a className={`button button-primary amb-join-link ${className}`} href={joinURL}>{label}<HugeiconsIcon icon={ArrowUpRight01Icon} size={18} strokeWidth={1.6} aria-hidden="true" /></a>;
}

export default function AmbassadorsPage() {
  return <div className="ambassadors-page">
    <section className="amb-hero" aria-labelledby="amb-hero-title">
      <AmbassadorsHeroArt />
      <div className="amb-hero-copy">
        <h1 id="amb-hero-title">Be the reason<br /><span className="amb-hero-gradient">people meet.</span></h1>
        <p className="amb-lead">Become a Hackways ambassador. Turn a shared interest into an event, a conversation, a community that keeps coming back.</p>
        <div className="amb-hero-actions"><JoinLink label="Become an ambassador" /><SectionJumpButton targetId="ambassador-role" className="amb-role-link">Explore the program<HugeiconsIcon icon={ArrowDown01Icon} size={17} strokeWidth={1.5} aria-hidden="true" /></SectionJumpButton></div>
      </div>
    </section>

    <section className="amb-manifesto site-container" aria-labelledby="amb-manifesto-title"><h2 id="amb-manifesto-title">Every community needs<br /><span>someone who starts.</span></h2><div><p>Someone to share the first invitation. Someone to bring an idea into the room. Someone to make a new face feel welcome.</p><p>If that sounds like you, bring your people to Hackways. You don’t need a huge audience to begin.</p></div></section>

    <section id="ambassador-role" className="amb-role site-container" aria-labelledby="amb-role-title" tabIndex={-1}>
      <div className="amb-section-intro"><h2 id="amb-role-title">From a shared interest<br />to a shared experience.</h2><p>Your role is to host your events through Hackways. Bring the idea, invite your community and use the platform to take care of the registration details.</p><Link className="amb-text-link" href="/organizers">See how hosting works<HugeiconsIcon icon={ArrowUpRight01Icon} size={17} strokeWidth={1.5} aria-hidden="true" /></Link></div>
      <ol className="amb-role-list" role="list">{responsibilities.map(item => <li key={item.title}><div><h3>{item.title}</h3><p>{item.description}</p></div></li>)}</ol>
    </section>

    <section className="amb-journey site-container" aria-labelledby="amb-journey-title"><div className="amb-section-intro"><h2 id="amb-journey-title">Your next chapter<br />starts with an idea.</h2><p>A thoughtful introduction tells us more than a polished pitch. Your application has three short steps, and you can save it along the way.</p><JoinLink label="Start your application" /></div><ol role="list"><li><span aria-hidden="true">1</span><div><h3>Introduce yourself.</h3><p>Share your LinkedIn profile, WhatsApp number and a short introduction to your community.</p></div></li><li><span aria-hidden="true">2</span><div><h3>Tell us what you’d host.</h3><p>Describe an event idea, the people it would bring together and why you want to make it happen.</p></div></li><li><span aria-hidden="true">3</span><div><h3>Review, submit, stay in the loop.</h3><p>Check your answers before submitting. We will review your profile and send updates by email. Results will be shared at the end of November.</p></div></li></ol></section>

    <section className="amb-open-section" aria-labelledby="amb-open-title"><div className="site-container amb-open-inner"><h2 id="amb-open-title">Your community.<br />Your kind of gathering.</h2><div><p>A workshop for people learning together. A meetup around a shared interest. A hackathon where an idea becomes a project. Start with the people you want to bring together.</p><p>You choose the gathering and organize the experience. Use Hackways to publish the details and manage registration.</p></div></div></section>

    <section className="amb-recognition site-container" aria-labelledby="amb-recognition-title">
      <div className="amb-section-intro"><h2 id="amb-recognition-title">Your contribution<br />gets recognized.</h2><p>This program starts with the work you do for your community. Here’s the recognition available today.</p></div>
      <dl className="amb-recognition-list"><div><dt>A letter of appreciation.</dt><dd>Recognition of the work you put into bringing people together as a Hackways ambassador.</dd></div><div><dt>Swag for top contributors.</dt><dd>The highest contributors can receive Hackways swag. It’s recognition for contribution, rather than a guaranteed joining reward.</dd></div></dl>
    </section>

    <section className="amb-faq site-container" aria-labelledby="amb-faq-title"><h2 id="amb-faq-title">A few things<br />you might be wondering.</h2><div className="amb-faq-list">{questions.map(item => <details key={item.question}><summary>{item.question}<span className="amb-faq-marker" aria-hidden="true" /></summary><p>{item.answer}</p></details>)}</div></section>

    <section id="ambassador-join" className="amb-join-section site-container" aria-labelledby="amb-join-title" tabIndex={-1}>
      <div><h2 id="amb-join-title">You bring the people.<br />Let’s make it happen.</h2><p>Your community might be one invitation away. Tell us about the gathering you want to create.</p><JoinLink label="Become an ambassador" /><p className="amb-application-note">Save your progress. Submit when you’re ready. Follow your result in your account.</p></div>
      <aside className="amb-joining-details" aria-labelledby="amb-application-title"><h3 id="amb-application-title">Come with an idea.<br />Start with an introduction.</h3><ul><li>Your profile and community</li><li>Your LinkedIn profile and WhatsApp number</li><li>Your event idea and motivation</li></ul><a className="amb-status-link" href={`${site.app}/programs/ambassadors/application`}>Already applied? View your application</a><p className="amb-privacy-note">Your submitted application is visible to you and the program review team. Read our <Link href="/privacy">privacy notice</Link>.</p></aside>
    </section>
  </div>;
}
