"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { site } from "@/lib/site";
import { Brand } from "./Brand";
import { AmbassadorsHeader } from '@/components/ambassadors/AmbassadorsHeader';

export function Header() {
  const [open, setOpen] = useState(false);
  const pathname = usePathname();
  const toggle = useRef<HTMLButtonElement>(null);

  useEffect(() => { setOpen(false); }, [pathname]);

  if (pathname === '/ambassadors' || pathname.startsWith('/ambassadors/')) return <AmbassadorsHeader />;

  return (
    <header className="site-header" onKeyDown={(event) => {
      if (event.key === "Escape" && open) {
        setOpen(false);
        toggle.current?.focus();
      }
    }}>
      <div className="site-container header-inner">
        <Brand />
        <button ref={toggle} className="menu-toggle" type="button" aria-expanded={open} aria-controls="site-navigation" onClick={() => setOpen(!open)}>
          {open ? "Close" : "Menu"}
        </button>
        <nav id="site-navigation" aria-label="Main navigation" className={`site-nav${open ? " is-open" : ""}`} onClick={(event) => {
          if (event.target instanceof Element && event.target.closest("a")) setOpen(false);
        }}>
          <Link href="/" aria-current={pathname === "/" ? "page" : undefined}>Discover</Link>
          <a href={`${site.app}/discover`}>Explore</a>
          <a href="/create">Create</a>
          <Link href="/ambassadors" aria-current={pathname === "/ambassadors" ? "page" : undefined}>Ambassadors</Link>
          <a className="button button-primary mobile-host-action" href="/create">Host an event</a>
        </nav>
        <a className="button button-primary header-action" href="/create">Host an event</a>
      </div>
    </header>
  );
}
