import Link from "next/link";

export function Brand() {
  return (
    <Link className="brand" href="/" aria-label="Hackways home">
      <span>hackways<span className="brand-dot">.</span></span>
    </Link>
  );
}
