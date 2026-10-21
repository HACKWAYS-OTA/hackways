'use client';

import { useEffect, useRef, type ReactNode } from 'react';

/** Pins the existing elements after the page reaches their reading position. */
export function EventScrollLayout({ children }: { children: ReactNode }) {
  const element = useRef<HTMLElement>(null);

  useEffect(() => {
    const root = element.current;
    if (!root) return;
    const titleSlot = root.querySelector<HTMLElement>('.event-title-anchor');
    const title = titleSlot?.querySelector<HTMLElement>('h1');
    const planSlot = root.querySelector<HTMLElement>('.event-plan');
    const plan = root.querySelector<HTMLElement>('.event-plan-anchor');
    const story = root.querySelector<HTMLElement>('.event-story');
    if (!titleSlot || !title || !planSlot || !plan || !story) return;
    const scrollingContent = Array.from(root.querySelectorAll<HTMLElement>('.event-title-block > :not(.event-title-anchor), .event-story')).map(node => ({ node, top: 0 }));

    const desktop = window.matchMedia('(min-width: 861px)');
    let frame = 0;
    let disposed = false;
    let threshold = 0;
    let titleTop = 0;
    let planHeight = 0;
    let titleHeight = 0;
    let footerTop = Number.POSITIVE_INFINITY;
    let contentTop = 0;

    function update() {
      frame = 0;
      if (!desktop.matches) {
        root!.removeAttribute('data-reading-pinned');
        return;
      }
      const scroll = window.scrollY;
      root!.toggleAttribute('data-reading-pinned', scroll >= threshold);
      // Both held elements must leave before the footer enters their space.
      const boundary = footerTop - scroll - 24;
      root!.style.setProperty('--event-plan-top', `${Math.min(40, boundary - planHeight)}px`);
      const heldTitleTop = Math.min(titleTop, boundary - titleHeight);
      root!.style.setProperty('--event-title-top', `${heldTitleTop}px`);
      contentTop = heldTitleTop + titleHeight + 16;
      // Clip only the moving content, leaving the shared theme visible behind the title.
      for (const item of scrollingContent) item.node.style.setProperty('--event-content-clip', `${Math.max(0, contentTop - (item.top - scroll))}px`);
    }

    function schedule() {
      if (!frame) frame = window.requestAnimationFrame(update);
    }

    function measure() {
      if (disposed) return;
      if (!desktop.matches) { update(); return; }
      const scroll = window.scrollY;
      const planBox = planSlot!.getBoundingClientRect();
      const titleBox = titleSlot!.getBoundingClientRect();
      planHeight = plan!.getBoundingClientRect().height;
      titleHeight = title!.getBoundingClientRect().height;
      for (const item of scrollingContent) item.top = item.node.getBoundingClientRect().top + scroll;
      threshold = Math.max(0, planBox.top + scroll - 40);
      titleTop = titleBox.top + scroll - threshold;
      root!.style.setProperty('--event-plan-left', `${planBox.left}px`);
      root!.style.setProperty('--event-plan-width', `${planBox.width}px`);
      root!.style.setProperty('--event-plan-height', `${planHeight}px`);
      root!.style.setProperty('--event-title-left', `${titleBox.left}px`);
      root!.style.setProperty('--event-title-width', `${titleBox.width}px`);
      root!.style.setProperty('--event-title-height', `${titleHeight}px`);
      const footer = document.querySelector<HTMLElement>('body > .app-footer');
      footerTop = footer ? footer.getBoundingClientRect().top + scroll : root!.getBoundingClientRect().bottom + scroll;
      update();
    }

    function revealFocused(event: FocusEvent) {
      if (!desktop.matches || !root!.hasAttribute('data-reading-pinned') || !(event.target instanceof HTMLElement)) return;
      const target = event.target;
      if (!scrollingContent.some(item => item.node.contains(target))) return;
      const top = target.getBoundingClientRect().top;
      if (top < contentTop + 6) window.scrollBy({ top: top - contentTop - 6, behavior: 'instant' });
    }

    const observer = new ResizeObserver(measure);
    observer.observe(titleSlot);
    observer.observe(title);
    observer.observe(planSlot);
    observer.observe(plan);
    observer.observe(story);
    const introduction = root.querySelector<HTMLElement>('.event-title-block');
    if (introduction) observer.observe(introduction);
    root.addEventListener('focusin', revealFocused);
    window.addEventListener('scroll', schedule, { passive: true });
    window.addEventListener('resize', measure);
    desktop.addEventListener('change', measure);
    void document.fonts.ready.then(measure);
    measure();

    return () => {
      disposed = true;
      observer.disconnect();
      root.removeEventListener('focusin', revealFocused);
      window.removeEventListener('scroll', schedule);
      window.removeEventListener('resize', measure);
      desktop.removeEventListener('change', measure);
      if (frame) window.cancelAnimationFrame(frame);
      root.removeAttribute('data-reading-pinned');
      for (const item of scrollingContent) item.node.style.removeProperty('--event-content-clip');
    };
  }, []);

  return <article ref={element} className="event-presentation">{children}</article>;
}
