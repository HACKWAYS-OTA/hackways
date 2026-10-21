"use client";

import { Children, useEffect, useRef, useState, type ReactNode } from "react";

export function EventStories({ children }: { children: ReactNode }) {
  const track = useRef<HTMLDivElement>(null);
  const [active, setActive] = useState(0);
  const count = Children.count(children);
  const stride = useRef(0);
  const current = useRef(0);

  useEffect(() => {
    const element = track.current;
    const first = element?.firstElementChild;
    if (!element || !first) return;
    function measure() {
      if (element && first) stride.current = first.getBoundingClientRect().width + (parseFloat(getComputedStyle(element).columnGap) || 0);
    }
    const resize = typeof ResizeObserver === "undefined" ? undefined : new ResizeObserver(measure);
    resize?.observe(element);
    measure();
    window.addEventListener("resize", measure);
    return () => { resize?.disconnect(); window.removeEventListener("resize", measure); };
  }, []);

  function moveTo(index: number) {
    const target = track.current?.children[index];
    target?.scrollIntoView({ inline: "start", block: "nearest", behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "instant" : "smooth" });
  }

  return (
    <div className="event-stories">
      <div className="site-container event-stories-shell">
        <div ref={track} className="event-story-track" role="group" aria-label="Event introductions" tabIndex={0} onScroll={(event) => {
          if (!stride.current) return;
          const index = Math.max(0, Math.min(count - 1, Math.round(event.currentTarget.scrollLeft / stride.current)));
          if (index !== current.current) { current.current = index; setActive(index); }
        }} onKeyDown={(event) => {
          if (event.target !== event.currentTarget || !["ArrowLeft", "ArrowRight"].includes(event.key)) return;
          event.preventDefault();
          moveTo(Math.max(0, Math.min(count - 1, active + (event.key === "ArrowRight" ? 1 : -1))));
        }}>{children}</div>
        <div className="event-story-controls">
          <button className="button button-outline" type="button" aria-label="Previous event introduction" disabled={active === 0} onClick={() => moveTo(active - 1)}>Previous</button>
          <span aria-live="polite" aria-atomic="true">{active + 1} of {count}</span>
          <button className="button button-outline" type="button" aria-label="Next event introduction" disabled={active === count - 1} onClick={() => moveTo(active + 1)}>Next</button>
        </div>
      </div>
    </div>
  );
}
