"use client";

import { useId, useRef, useState } from "react";
import { HugeiconsIcon } from "@hugeicons/react";
import BookOpen01Icon from "@hugeicons/core-free-icons/BookOpen01Icon";
import CodeIcon from "@hugeicons/core-free-icons/CodeIcon";
import Globe02Icon from "@hugeicons/core-free-icons/Globe02Icon";
import LaptopIcon from "@hugeicons/core-free-icons/LaptopIcon";
import Mic01Icon from "@hugeicons/core-free-icons/Mic01Icon";
import UserGroupIcon from "@hugeicons/core-free-icons/UserGroupIcon";

const icons = { book: BookOpen01Icon, code: CodeIcon, globe: Globe02Icon, laptop: LaptopIcon, mic: Mic01Icon, people: UserGroupIcon };

export type ContentChoice = {
  label: string;
  title: string;
  description: string;
  note?: string;
  icon: keyof typeof icons;
};

export function ContentSwitcher({ label, items }: { label: string; items: readonly ContentChoice[] }) {
  const [selected, setSelected] = useState(0);
  const id = useId();
  const tabs = useRef<(HTMLButtonElement | null)[]>([]);

  return (
    <div className="content-switcher">
      <div className="content-tabs" role="tablist" aria-label={label}>
        {items.map((item, index) => (
          <button
            key={item.label}
            ref={(element) => { tabs.current[index] = element; }}
            type="button"
            role="tab"
            id={`${id}-tab-${index}`}
            aria-controls={`${id}-panel-${index}`}
            aria-selected={selected === index}
            tabIndex={selected === index ? 0 : -1}
            onClick={() => setSelected(index)}
            onKeyDown={(event) => {
              const next = event.key === "ArrowRight" ? (index + 1) % items.length
                : event.key === "ArrowLeft" ? (index - 1 + items.length) % items.length
                : event.key === "Home" ? 0 : event.key === "End" ? items.length - 1 : undefined;
              if (next === undefined) return;
              event.preventDefault();
              setSelected(next);
              tabs.current[next]?.focus();
            }}
          >{item.label}</button>
        ))}
      </div>
      {items.map((item, index) => (
        <div key={item.label} className="content-panel" role="tabpanel" id={`${id}-panel-${index}`} aria-labelledby={`${id}-tab-${index}`} hidden={selected !== index} tabIndex={0}>
          <div className={`content-panel-icon content-panel-icon--${item.icon}`} aria-hidden="true"><HugeiconsIcon data-draw-icon icon={icons[item.icon]} size={40} strokeWidth={1.5} /></div>
          <div className="content-panel-copy">
            <h3>{item.title}</h3>
            <p>{item.description}</p>
            {item.note && <details className="context-note"><summary>Good to know<span aria-hidden="true" /></summary><p>{item.note}</p></details>}
          </div>
        </div>
      ))}
    </div>
  );
}
