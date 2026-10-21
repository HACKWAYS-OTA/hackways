"use client";
export function SkipToContent() {
  return <button className="skip-control" type="button" onClick={() => { const main = document.getElementById('main-content'); main?.focus({ preventScroll: true }); main?.scrollIntoView({ block: 'start' }); }}>Skip to content</button>;
}
