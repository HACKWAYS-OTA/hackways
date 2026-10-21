"use client";

import { useEffect, useRef, useState } from "react";
import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";

const destinations = [
  ["Events", "hackathons-heading"],
  ["Attending", "rsvp-heading"],
  ["Hosting", "hosting-heading"],
  ["Questions", "answers-heading"],
] as const;

export function LandingNavigation() {
  const navigation = useRef<HTMLElement>(null);
  const [current, setCurrent] = useState<string>(destinations[0][1]);

  useEffect(() => {
    gsap.registerPlugin(ScrollTrigger);
    const context = gsap.context(() => {
      destinations.forEach(([, target], index) => {
        const heading = document.getElementById(target);
        if (!heading) return;
        ScrollTrigger.create({
          trigger: heading,
          start: () => `top ${parseFloat(getComputedStyle(document.documentElement).getPropertyValue("--header-height")) + (navigation.current?.offsetHeight ?? 60) + 80}px`,
          onEnter: () => setCurrent(target),
          onLeaveBack: () => setCurrent(destinations[Math.max(0, index - 1)][1]),
          onRefresh: (trigger) => { if (trigger.progress > 0) setCurrent(target); },
        });
      });
      ScrollTrigger.create({
        trigger: ".landing-page", start: "top top", end: "bottom bottom",
        onUpdate: ({ progress }) => navigation.current?.style.setProperty("--page-progress", String(progress)),
      });
    });
    return () => context.revert();
  }, []);

  return (
    <nav ref={navigation} className="landing-jump-nav" aria-label="On this page">
      <div className="site-container">
        {destinations.map(([label, target]) => <button key={target} type="button" aria-current={current === target ? "location" : undefined} onClick={() => {
          const heading = document.getElementById(target);
          if (!heading) return;
          heading.focus({ preventScroll: true });
          heading.scrollIntoView({ block: "start", behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "instant" : "smooth" });
        }}>{label}</button>)}
      </div>
    </nav>
  );
}
