import type { ReactNode } from "react";
import './article.css';

export function Article({ title, intro, variant = 'policy', actions, updated, children }: {
  title: string;
  intro: string;
  variant?: 'policy' | 'about' | 'organizers' | 'help' | 'contact';
  actions?: ReactNode;
  updated?: string;
  children: ReactNode;
}) {
  return (
    <article className={`article-page information-page information-${variant}`}>
      <header className="article-header">
        <h1 id="article-title" tabIndex={-1}>{title}</h1>
        <div className="article-introduction">
          <p>{intro}</p>
          {actions && <div className="section-actions">{actions}</div>}
          {updated && <p className="article-updated">Updated <time dateTime="2026-10-20">{updated}</time></p>}
        </div>
      </header>
      <div className="article-body">{children}</div>
    </article>
  );
}

export function ArticleSection({ id, title, children }: { id: string; title: string; children: ReactNode }) {
  return <section id={id} className="article-section" aria-labelledby={`${id}-heading`}><h2 id={`${id}-heading`} tabIndex={-1}>{title}</h2><div className="article-section-copy">{children}</div></section>;
}
