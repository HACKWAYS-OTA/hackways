"use client";

import { useEffect } from "react";
import { scrollToSection } from '@hackways/ui/section-jump';

export function SkipToContent() {
  useEffect(() => {
    function removeLegacyFragment() { if (window.location.hash) {
      let id = '';
      try { id = decodeURIComponent(window.location.hash.slice(1)); } catch { /* An invalid legacy fragment still gets removed. */ }
      window.history.replaceState(window.history.state, "", window.location.pathname + window.location.search);
      if (id) scrollToSection(document.getElementById(`${id}-heading`) ? `${id}-heading` : id);
    } }
    removeLegacyFragment();
    window.addEventListener('hashchange', removeLegacyFragment);
    return () => window.removeEventListener('hashchange', removeLegacyFragment);
  }, []);

  function moveToContent() {
    scrollToSection(document.getElementById('article-title') ? 'article-title' : 'main-content');
  }

  return <button type="button" className="skip-link" onClick={moveToContent}>Skip to content</button>;
}
