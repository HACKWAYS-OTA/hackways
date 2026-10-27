import type { ReactNode } from 'react';
import Link from 'next/link';
import { site } from '@/lib/site';
import { CookiePreferencesButton } from '@hackways/ui/analytics';
export function ProgramFrame({ children, title, section, slug, directory = false }: { children: ReactNode; title: string; section: string; slug?: string; directory?: boolean }) {
  const ambassador = slug === 'ambassadors';
  return <div id="program-surface" data-program={slug} data-view={directory ? 'directory' : undefined}>
    <header className="program-header">
      <a className="program-wordmark" href={ambassador ? `${site.public}/ambassadors` : site.public}>hackways.</a>
      <nav aria-label="Program navigation">
        {!ambassador && <Link href="/programs">Programs</Link>}
        {slug && <Link href={`/programs/${slug}/application`}>Your application</Link>}
        <Link href="/account">Account</Link>
      </nav>
    </header>
    <main id="main-content" className="program-main" tabIndex={-1}>
      <div className="program-page-intro">
        {!directory && <nav aria-label="Breadcrumb" className="program-context">
          {ambassador ? <a href={`${site.public}/ambassadors`}>Ambassadors</a> : <Link href="/programs">Programs</Link>}
          <span aria-hidden="true"> / </span><span>{section}</span>
        </nav>}
        <h1><span>{title}</span></h1>
      </div>
      {children}
    </main>
    <footer className="program-footer"><span>Hackways</span><div><CookiePreferencesButton /><a href={`${site.public}/privacy`}>Privacy</a><a href={`${site.public}/terms`}>Terms</a><a href={`${site.public}/ambassadors`}>About the program</a></div></footer>
  </div>;
}
