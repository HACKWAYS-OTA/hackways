"use client";

import { useEffect } from "react";
import { usePathname } from "next/navigation";
import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";

export function SiteMotion() {
  const pathname = usePathname();

  useEffect(() => {
    const root = document.getElementById("main-content");
    if (!root) return;
    gsap.registerPlugin(ScrollTrigger);
    const media = gsap.matchMedia(root);
    let live = true;
    let refresh: gsap.core.Tween | undefined;

    function compose(context: gsap.Context) {
      if (!root) return;
      let observing = true;
      const desktop = window.matchMedia("(min-width: 901px)").matches;

      const mosaic = root.querySelector<HTMLElement>(".interest-mosaic");
      if (mosaic) {
        const assembly = gsap.timeline({ scrollTrigger: { trigger: mosaic, start: "top 88%", end: "center 55%", scrub: true, invalidateOnRefresh: true } });
        mosaic.querySelectorAll<HTMLElement>(".interest-tile--icon").forEach((tile, index) => {
          const compact = window.matchMedia("(max-width: 680px)").matches;
          const column = Number(tile.style.getPropertyValue(compact ? "--compact-column" : "--tile-column"));
          assembly.fromTo(tile, { x: (column - (compact ? 4 : 7)) * (desktop ? 9 : 4), y: desktop ? 72 : 32, rotation: (index % 2 ? 1 : -1) * (desktop ? 7 : 3), scale: .88 }, { x: 0, y: 0, rotation: 0, scale: 1, duration: 1, ease: "power3.out" }, (index % 5) * .07);
        });
      }

      // One observer also catches icons revealed by tabs or horizontal swiping.
      // Keep the server-rendered icon visible until its animation actually starts.
      const icons = typeof IntersectionObserver === "undefined" ? undefined : new IntersectionObserver((entries) => {
        if (!observing) return;
        entries.forEach((entry) => {
          if (!entry.isIntersecting) return;
          icons?.unobserve(entry.target);
          context.add(() => {
            const drawing = gsap.timeline();
            entry.target.querySelectorAll<SVGGeometryElement>("path, circle, rect, line, polyline, polygon, ellipse").forEach((shape, index) => {
              const length = shape.getTotalLength();
              if (!Number.isFinite(length) || length <= 0 || getComputedStyle(shape).stroke === "none") return;
              drawing.fromTo(shape, { strokeDasharray: length, strokeDashoffset: length }, { strokeDashoffset: 0, duration: .8, ease: "power2.inOut" }, index * .05);
            });
          });
        });
      }, { threshold: .2, rootMargin: "0px 0px -14% 0px" });
      root.querySelectorAll<SVGSVGElement>("svg[data-draw-icon]").forEach((icon) => icons?.observe(icon));
      return () => { observing = false; icons?.disconnect(); };
    }

    media.add("(min-width: 901px) and (prefers-reduced-motion: no-preference)", compose);
    media.add("(max-width: 900px) and (prefers-reduced-motion: no-preference)", compose);
    function contentChanged() {
      refresh?.kill();
      refresh = gsap.delayedCall(.12, () => ScrollTrigger.refresh());
    }
    const resize = typeof ResizeObserver === "undefined" ? undefined : new ResizeObserver(contentChanged);
    resize?.observe(root);
    root.addEventListener("toggle", contentChanged, true);
    document.fonts.ready.then(() => { if (live) ScrollTrigger.refresh(); });
    return () => {
      live = false;
      refresh?.kill();
      resize?.disconnect();
      root.removeEventListener("toggle", contentChanged, true);
      media.revert();
    };
  }, [pathname]);

  return null;
}
