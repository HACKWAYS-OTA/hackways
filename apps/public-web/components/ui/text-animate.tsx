"use client";

// Adapted from the user-supplied Magic UI Text Animate by dillionverma.
import { memo, useEffect, useMemo, useRef } from "react";
import { motion, useAnimationControls, useInView, useReducedMotion, type MotionProps, type Variants } from "motion/react";

const elements = {
  article: motion.article, div: motion.div, h1: motion.h1, h2: motion.h2,
  h3: motion.h3, h4: motion.h4, h5: motion.h5, h6: motion.h6,
  li: motion.li, p: motion.p, section: motion.section, span: motion.span,
} as const;

type AnimationType = "text" | "word" | "character" | "line";
type AnimationVariant = "fadeIn" | "blurIn" | "blurInUp" | "blurInDown" | "slideUp" | "slideDown" | "slideLeft" | "slideRight" | "scaleUp" | "scaleDown";

interface TextAnimateProps extends Omit<MotionProps, "children"> {
  children: string;
  id?: string;
  className?: string;
  segmentClassName?: string;
  delay?: number;
  duration?: number;
  variants?: Variants;
  as?: keyof typeof elements;
  by?: AnimationType;
  startOnView?: boolean;
  once?: boolean;
  animation?: AnimationVariant;
  accessible?: boolean;
}

const starts: Record<AnimationVariant, Variants["hidden"]> = {
  fadeIn: { opacity: 0 },
  blurIn: { opacity: 0, filter: "blur(4px)" },
  blurInUp: { opacity: 0, filter: "blur(4px)", y: 12 },
  blurInDown: { opacity: 0, filter: "blur(4px)", y: -12 },
  slideUp: { opacity: 0, y: 16 },
  slideDown: { opacity: 0, y: -16 },
  slideLeft: { opacity: 0, x: 16 },
  slideRight: { opacity: 0, x: -16 },
  scaleUp: { opacity: 0, scale: .85 },
  scaleDown: { opacity: 0, scale: 1.15 },
};

function TextAnimateBase({ children, as = "p", by = "word", animation = "fadeIn", delay = 0, duration = .3, variants, className = "", segmentClassName = "", startOnView = true, once = false, accessible = true, ...props }: TextAnimateProps) {
  const Component = elements[as];
  const ref = useRef<HTMLElement | null>(null);
  const played = useRef(false);
  const controls = useAnimationControls();
  const reduced = useReducedMotion();
  const inView = useInView(ref, { once, amount: .2 });
  const shouldStart = !startOnView || inView;
  const segments = useMemo(() => by === "word" ? children.split(/(\s+)/) : by === "character" ? Array.from(children) : by === "line" ? children.split("\n") : [children], [by, children]);
  const count = Math.max(1, segments.filter((segment) => segment.trim()).length);
  const container: Variants = {
    hidden: {},
    show: { transition: { delayChildren: delay, staggerChildren: duration / count } },
  };
  const item: Variants = variants ?? {
    hidden: starts[animation],
    show: { opacity: 1, x: 0, y: 0, scale: 1, filter: "blur(0px)", transition: { duration: Math.max(.4, duration * .8), ease: [.16, 1, .3, 1] } },
  };

  useEffect(() => {
    // Server HTML stays readable. Animate only after the preference is known.
    if (reduced !== false) { controls.set("show"); return; }
    if (!shouldStart) return;
    if (once && played.current) { controls.set("show"); return; }
    let active = true;
    controls.set("hidden");
    void controls.start("show").then(() => { if (active) played.current = true; });
    return () => { active = false; controls.stop(); };
  }, [controls, reduced, shouldStart, once, children, animation, by]);

  return (
    <Component {...props} ref={(node: HTMLElement | null) => { ref.current = node; }} variants={container} initial={false} animate={controls} className={`text-animate ${className}`} aria-label={accessible ? children.replace(/\s+/g, " ") : undefined}>
      {segments.map((segment, index) => !segment.trim() ? segment : (
        <motion.span key={`${by}-${index}`} variants={item} className={`text-animate-segment${by === "line" ? " text-animate-line" : ""} ${segmentClassName}`} aria-hidden={accessible || undefined}>
          {segment}
        </motion.span>
      ))}
    </Component>
  );
}

export const TextAnimate = memo(TextAnimateBase);
