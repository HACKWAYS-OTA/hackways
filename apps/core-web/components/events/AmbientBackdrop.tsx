'use client';

import { useEffect, useState } from 'react';

function publicImage(source?: string | null): string {
  try {
    const url = new URL(source || '');
    return url.protocol === 'https:' && !url.username && !url.password ? url.href : '';
  } catch { return ''; }
}

/** Decorative colour wash from the real cover, never an invented theme image. */
export function AmbientBackdrop({ active, file, source }: { active: boolean; file?: File | null; source?: string | null }) {
  const [preview, setPreview] = useState('');
  useEffect(() => {
    setPreview('');
    if (!active || !file) return;
    let current = true;
    const fallback = URL.createObjectURL(file);
    // Keep the preview texture small; no full-size image processing on scroll.
    async function prepare() {
      try {
        const bitmap = await createImageBitmap(file!, { resizeWidth: 96, resizeHeight: 96, resizeQuality: 'low' });
        try {
          if (!current) return;
          const canvas = document.createElement('canvas');
          canvas.width = 96; canvas.height = 96;
          const context = canvas.getContext('2d');
          if (!context) { setPreview(fallback); return; }
          context.drawImage(bitmap, 0, 0, 96, 96);
          setPreview(canvas.toDataURL('image/jpeg', 0.65));
        } finally { bitmap.close(); }
      } catch { if (current) setPreview(fallback); }
    }
    void prepare();
    return () => { current = false; URL.revokeObjectURL(fallback); };
  }, [active, file]);
  const image = file ? preview : publicImage(source);
  return active && image ? <div className="ambient-backdrop" aria-hidden="true" style={{ backgroundImage: `url(${JSON.stringify(image)})` }} /> : null;
}
