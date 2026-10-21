import type { Metadata } from 'next';
import Link from 'next/link';
import { Article, ArticleSection as Section } from '@/components/site/Article';
import { supportEmail } from '@/lib/site';
export const metadata: Metadata = { title: 'Acceptable use', description: 'Guidelines for honest event listings, respectful participation, guest privacy, responsible registration and reporting misuse on Hackways.', alternates: { canonical: '/acceptable-use' } };
export default function AcceptableUsePage() {
  return <Article title="Acceptable use" intro="Keep event information truthful, participation respectful and private information protected. These expectations apply to listings, profiles, registration answers, notes and the ways people use Hackways." updated="20 October 2026">
    <Section id="honesty" title="Publish real events and represent yourself accurately">
      <p>Use a genuine identity or community name you are entitled to represent. Describe the programme, venue, schedule, eligibility, rewards and affiliations accurately. Correct material errors when plans change.</p>
      <ul><li>Do not impersonate another person, organizer, institution or business.</li><li>Do not publish fabricated events, deceptive prizes, false endorsements or misleading admission claims.</li><li>Do not use an event listing to disguise unrelated solicitation, fraud or an attempt to obtain credentials.</li><li>Do not describe a private link as an access restriction or an RSVP as confirmed when it is only an application.</li></ul>
    </Section>
    <Section id="respect" title="Treat other people with care">
      <p>Do not use Hackways to threaten, harass, intimidate or target people with discriminatory abuse. Do not publish material that exposes someone to harm or encourages others to do so.</p>
      <p>Respect event-specific participation rules, accessibility needs and personal boundaries. A disagreement about a programme or registration decision does not justify abusive messages, repeated unwanted contact or public exposure of personal information.</p>
    </Section>
    <Section id="privacy" title="Keep private information private">
      <p>Guest names, email addresses, answers, notes and ticket details are not a public directory. Organizers should use the guest information available to them to administer the event and keep it secure.</p>
      <ul><li>Do not publish another guest list, contact details, registration answers or private messages without permission.</li><li>Do not share passwords, verification codes, session cookies or usable ticket QR codes.</li><li>Do not sell guest details or use them for unsolicited bulk messages.</li><li>Do not collect sensitive information through an event form unless it is necessary, clearly explained and appropriate for the event.</li></ul>
      <p>If you find exposed personal information, report its location privately and avoid copying or redistributing it.</p>
    </Section>
    <Section id="registration" title="Give other guests a fair chance to join">
      <p>Register using accurate information and use only the actions available for your booking. Respect capacity, approval and waiting-list decisions.</p>
      <p>Do not create fabricated or duplicate accounts to hold places, bypass a rejection, skip a waiting list or evade a restriction. Do not sell or distribute registrations through misleading claims. Use a supported transfer action where one is available rather than forwarding a ticket screenshot.</p>
      <p>If you cannot attend, cancel an eligible registration so the place can become available to someone else. A waiting-list entry, expired offer or pending application does not entitle someone to enter an event.</p>
    </Section>
    <Section id="security" title="Use technical access responsibly">
      <p>Do not attempt to access another account, retrieve private records, bypass authorization or capacity controls, forge tickets or interfere with check-in. Do not distribute malware, phishing links or content designed to compromise another device.</p>
      <p>Do not overwhelm the service with automated requests, spam submissions or disruptive activity. Publicly viewable event information does not authorize access to private guest data or management endpoints.</p>
      <p>If you suspect a vulnerability, stop at the minimum evidence needed to describe it using your own records. Report it privately. Do not download unrelated data, test against other accounts or disrupt a live event to prove the issue.</p>
    </Section>
    <Section id="content" title="Use content you have permission to publish">
      <p>Only upload covers, logos and other material you own or are allowed to use. Respect intellectual-property rights and the permissions attached to third-party material.</p>
      <p>Do not publish illegal content, exploitative material, instructions intended to facilitate harm, or links disguised to mislead visitors. Clearly identify external destinations and explain any event-specific requirement before guests register.</p>
      <p>For a concern about your work being used without permission, send the affected page link, identify the material and explain your relationship to it. Do not include unrelated private records.</p>
    </Section>
    <Section id="report" title="Report a concern with useful facts">
      <p>Email <a href={'mailto:' + supportEmail}>{supportEmail}</a> with the event or page link, a factual description and the approximate time. Include a limited screenshot when it helps, with private information cropped out.</p>
      <p>Reports need context and assessment. Sending a report does not automatically establish a violation, guarantee removal or promise a particular resolution. Avoid speculative accusations and do not encourage others to target the person you are reporting.</p>
      <p>For an immediate threat to safety, contact local emergency services or venue security. Hackways support is not an emergency service.</p>
      <p>The <Link href="/contact">contact guide</Link> explains how to raise a concern. Read the <Link href="/terms">terms of use</Link> and <Link href="/privacy">privacy notice</Link> for related responsibilities and information handling.</p>
    </Section>
  </Article>;
}
