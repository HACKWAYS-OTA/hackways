'use client';

import { useEffect, useRef } from 'react';
import { site } from '@/lib/site';
import { Brand } from '@/components/site/Brand';
import './AmbassadorsHeader.css';

const applicationPath = '/programs/ambassadors/apply/profile';

export function AmbassadorsHeader() {
  const headerRef = useRef<HTMLElement>(null);
  useEffect(() => {
    const header = headerRef.current;
    if (!header) return;
    let frame = 0;
    const measure = () => {
      frame = 0;
      const hero = document.querySelector<HTMLElement>('.amb-hero');
      const rect = hero?.getBoundingClientRect();
      header.dataset.surface = !rect || rect.bottom <= header.offsetHeight + 20
        ? 'content' : rect.top < -rect.height * 0.48 ? 'dark' : 'light';
    };
    const schedule = () => { if (!frame) frame = requestAnimationFrame(measure); };
    measure();
    window.addEventListener('scroll', schedule, { passive: true });
    window.addEventListener('resize', schedule, { passive: true });
    window.addEventListener('pageshow', schedule);
    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener('scroll', schedule);
      window.removeEventListener('resize', schedule);
      window.removeEventListener('pageshow', schedule);
    };
  }, []);

  return <header ref={headerRef} className="ambassador-header" data-surface="light">
    <div className="ambassador-header-inner">
      <Brand />
      <nav aria-label="Ambassador navigation">
        <a className="ambassador-sign-in" href={`${site.app}/login?next=${encodeURIComponent(applicationPath)}`}>Sign in</a>
      </nav>
    </div>
  </header>;
}
