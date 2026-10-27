import type { CSSProperties } from "react";
import { HugeiconsIcon } from "@hugeicons/react";
import {
  ArrowRight01Icon,
  BookOpen01Icon,
  BulbIcon,
  Calendar03Icon,
  CodeIcon,
  Globe02Icon,
  LaptopIcon,
  Message01Icon,
  Mic01Icon,
  Rocket01Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import { site } from "@/lib/site";

// Coordinates describe decorative artwork, not event listings or category filters.
const iconTiles = [
  { row: 1, column: 13, icon: Globe02Icon, tone: "violet", compact: [1, 7] },
  { row: 2, column: 10, icon: BulbIcon, tone: "ochre", compact: [2, 4] },
  { row: 3, column: 8, icon: CodeIcon, tone: "ink", compact: [3, 2] },
  { row: 3, column: 12, icon: Mic01Icon, tone: "rose", compact: [3, 6] },
  { row: 3, column: 14, icon: UserGroupIcon, tone: "violet", compact: [2, 8] },
  { row: 4, column: 10, icon: BookOpen01Icon, tone: "violet", compact: [4, 4] },
  { row: 5, column: 4, icon: Rocket01Icon, tone: "rose", compact: [5, 2] },
  { row: 5, column: 12, icon: Message01Icon, tone: "ink", compact: [4, 8] },
  { row: 6, column: 3, icon: LaptopIcon, tone: "ink", compact: [5, 5] },
  { row: 6, column: 7, icon: Calendar03Icon, tone: "ochre", compact: [6, 6] },
] as const;

function cellsInRows(ranges: readonly (readonly [number, number, number])[]) {
  return ranges.flatMap(([row, first, last]) =>
    Array.from({ length: last - first + 1 }, (_, offset) => ({ row, column: first + offset })),
  );
}

function MosaicJoins({ cells: layout, columns, rows, compact = false }: {
  cells: readonly { row: number; column: number }[];
  columns: number;
  rows: number;
  compact?: boolean;
}) {
  const vertices = new Set(layout.flatMap(({ row, column }) => [
    `${column - 1},${row - 1}`, `${column},${row - 1}`,
    `${column - 1},${row}`, `${column},${row}`,
  ]));

  return (
    <svg className={`interest-joins${compact ? " interest-joins--compact" : " interest-joins--desktop"}`} viewBox={`0 0 ${columns * 100} ${rows * 100}`} fill="currentColor" focusable="false">
      {Array.from(vertices, (vertex) => {
        const [column, row] = vertex.split(",").map(Number);
        return (
          <g key={vertex} transform={`translate(${column * 100} ${row * 100})`}>
            <path d="M0-20A20 20 0 0 1-20 0A20 20 0 0 1 0 20A20 20 0 0 1 20 0A20 20 0 0 1 0-20Z" />
            <path d="M0-32V-20M-32 0H-20M20 0H32M0 20V32" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" vectorEffect="non-scaling-stroke" />
          </g>
        );
      })}
    </svg>
  );
}

const cells = cellsInRows([
  [1, 11, 13], [2, 8, 13], [3, 8, 14], [4, 5, 13],
  [5, 1, 13], [6, 1, 14], [7, 3, 7], [7, 10, 13],
]);
const compactCells = cellsInRows([
  [1, 6, 8], [2, 4, 8], [3, 2, 8], [4, 1, 8], [5, 1, 8], [6, 2, 7],
]);
const compactBackground = compactCells.filter(({ row, column }) => !iconTiles.some(({ compact }) => compact[0] === row && compact[1] === column));

export function InterestTiles() {
  let backgroundIndex = 0;

  return (
    <section className="landing-section interests-section" aria-labelledby="interests-heading">
      <div className="interest-composition">
        <div className="interest-copy">
          <h2 id="interests-heading">Find your people.<br />Follow your curiosity.</h2>
          <p>Explore AI, software, data, cloud, cybersecurity, startups and design. Find a gathering around the ideas you want to learn and build.</p>
          <a className="button button-outline" href={`${site.app}/discover`}>
            Explore events <HugeiconsIcon data-draw-icon icon={ArrowRight01Icon} size={18} aria-hidden="true" />
          </a>
        </div>
        <div className="interest-mosaic" aria-hidden="true">
          <MosaicJoins cells={cells} columns={14} rows={7} />
          <MosaicJoins cells={compactCells} columns={8} rows={6} compact />
          {cells.map(({ row, column }, index) => {
            const illustrated = iconTiles.find((tile) => tile.row === row && tile.column === column);
            const compact = illustrated
              ? { row: illustrated.compact[0], column: illustrated.compact[1] }
              : compactBackground[backgroundIndex++];
            const style = {
              "--tile-row": row,
              "--tile-column": column,
              "--compact-row": compact?.row,
              "--compact-column": compact?.column,
            } as CSSProperties;
            const tint = index % 13 === 0 ? "lavender" : index % 17 === 0 ? "cream" : index % 19 === 0 ? "mist" : "plain";

            return (
              <span
                key={`${row}-${column}`}
                className={`interest-tile${illustrated ? ` interest-tile--icon interest-tile--${illustrated.tone}` : ` interest-tile--${tint}`}${compact ? "" : " interest-tile--desktop-only"}`}
                style={style}
              >
                {illustrated && <HugeiconsIcon data-draw-icon icon={illustrated.icon} size={48} strokeWidth={1.5} />}
              </span>
            );
          })}
        </div>
      </div>
    </section>
  );
}
