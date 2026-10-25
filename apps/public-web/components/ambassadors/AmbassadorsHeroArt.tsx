'use client';

import { useEffect, useRef, useState } from 'react';
import { Toggle } from '@hackways/ui/toggle';
import GradientASCII from './GradientASCII';

export function AmbassadorsHeroArt() {
  const [paused, setPaused] = useState(false);
  const artRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const hero = artRef.current?.closest<HTMLElement>('.amb-hero');
    if (!hero) return;
    const preference = window.matchMedia('(prefers-reduced-motion: reduce)');
    let frame = 0;
    let current = 0;
    let target = 0;
    let lastTime = 0;

    const draw = (now: number) => {
      frame = 0;
      const elapsed = lastTime ? Math.min(now - lastTime, 64) : 16;
      lastTime = now;
      current += (target - current) * (1 - Math.exp(-elapsed / 125));
      if (Math.abs(target - current) < 0.0005) current = target;
      hero.style.setProperty('--amb-scroll-progress', current.toFixed(5));
      if (current !== target) frame = requestAnimationFrame(draw);
      else lastTime = 0;
    };
    const measure = (immediate = false) => {
      const distance = Math.max(240, Math.min(hero.offsetHeight * 0.6, window.innerHeight * 0.65));
      const progress = Math.min(1, Math.max(0, window.scrollY / distance));
      target = preference.matches ? 1 : progress * progress * (3 - 2 * progress);
      if (immediate || preference.matches) {
        cancelAnimationFrame(frame);
        frame = 0;
        lastTime = 0;
        current = target;
        hero.style.setProperty('--amb-scroll-progress', current.toFixed(5));
      } else if (!frame && !document.hidden) frame = requestAnimationFrame(draw);
    };
    const onScroll = () => measure();
    const onResize = () => measure(true);
    const onVisibility = () => {
      if (document.hidden) { cancelAnimationFrame(frame); frame = 0; lastTime = 0; }
      else measure(true);
    };
    measure(true);
    window.addEventListener('scroll', onScroll, { passive: true });
    window.addEventListener('resize', onResize, { passive: true });
    window.addEventListener('pageshow', onResize);
    preference.addEventListener('change', onResize);
    document.addEventListener('visibilitychange', onVisibility);
    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener('scroll', onScroll);
      window.removeEventListener('resize', onResize);
      window.removeEventListener('pageshow', onResize);
      preference.removeEventListener('change', onResize);
      document.removeEventListener('visibilitychange', onVisibility);
      hero.style.removeProperty('--amb-scroll-progress');
    };
  }, []);

  return <>
    <div ref={artRef} className="amb-hero-art" data-paused={paused ? 'true' : 'false'} aria-hidden="true">
      <GradientASCII
        colors={['#1e3fd8', '#f6f8ff', '#a5d8ff', '#38bdf8']}
        shape="horizon" characters=" =+xX0S8#@" cellSize={18} density={0.8}
        speed={paused ? 0 : 1} mouseInteraction={!paused}
        clickEffect={paused ? 'none' : 'wave'} flicker={paused ? 0 : 0.2}
        glow={0.2} grain={0.08} dpr={1.25} intro={false}
      />
    </div>
    <Toggle className="amb-motion-control" pressed={paused} onPressedChange={setPaused} aria-label="Pause background animation">
      {paused ? 'Resume animation' : 'Pause animation'}
    </Toggle>
  </>;
}
