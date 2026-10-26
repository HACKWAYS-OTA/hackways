import Image from 'next/image';
export function ProgramBanner({ slug, priority = false }: { slug: string; priority?: boolean }) {
  if (slug !== 'ambassadors') return null;
  return <Image className="program-banner" src="/programs/ambassadors.jpg" alt="Hackways Ambassadors, blue ASCII program banner" width={1931} height={1280} priority={priority} unoptimized sizes="(max-width: 560px) 88px, 144px" />;
}
