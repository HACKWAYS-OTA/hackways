import type { Metadata } from 'next';
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { Article, ArticleSection as Section } from '@/components/site/Article';
import { FAQ } from '@/components/site/FAQ';
import { faqTopics } from '@/lib/help-topics';
import { supportEmail } from '@/lib/site';
export function generateStaticParams() { return faqTopics.map(topic => ({ topic: topic.id })); }
export function generateMetadata({ params }: { params: { topic: string } }): Metadata {
  const topic = faqTopics.find(item => item.id === params.topic);
  if (!topic) notFound();
  return { title: topic.label, description: topic.intro, alternates: { canonical: `/help/${topic.id}` } };
}
export default function HelpTopic({ params }: { params: { topic: string } }) {
  const topic = faqTopics.find(item => item.id === params.topic);
  if (!topic) notFound();
  return <Article title={topic.label} intro={topic.intro}>
    <Section id="answers" title="Common questions"><FAQ topic={topic.id} /></Section>
    <Section id="support" title="Get more help"><p>Email <a href={`mailto:${supportEmail}`}>{supportEmail}</a> with the relevant event link and a description of the problem. Avoid including passwords, verification codes or private guest information.</p><p><Link href="/contact">Contact guidance</Link> explains what to send. You can also <Link href="/help">browse all help topics</Link>.</p></Section>
  </Article>;
}
