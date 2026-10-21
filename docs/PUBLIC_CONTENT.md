# Public information and support pages

The public-web routes `/about`, `/organizers`, `/help`, `/contact`, `/privacy`, `/terms`, `/cookies` and `/acceptable-use` use public website layouts. About groups attendee, host and community information; Organizers uses spacious heading/content rows; Help links to six topic pages and shared questions; Contact separates support details from event and reporting guidance. Policies use a simple reading column. There are no documentation sidebars, section indexes or secondary category toolbar. Essential content is server rendered without a client content fetch.

## Clean navigation paths

Do not generate fragment URLs. Dedicated help routes are `/help/finding`, `/help/joining`, `/help/changes`, `/help/accounts`, `/help/hosting` and `/help/safety`; they reuse the same FAQ content, provide canonical metadata and appear in the sitemap. Unknown topics return 404. Cross-page links use normal paths, including `/contact` for privacy/report guidance.

In-page controls scroll and focus headings through `@hackways/ui/section-jump` without changing the address or creating history entries. Section IDs exist for accessibility and DOM targeting, not fragment links. Existing public-site fragments are consumed and removed through history replacement; skip-to-content remains a button. Management registration shortcuts use the same shared jump button; management cross-page edit links use their clean editor/registration paths.

## Reading conventions

Use regular 400 weight with no synthetic heading stroke on these pages. Body copy is 17px, navigation and secondary text at least 16px, section headings 23–25px and primary headings 30–36px. Page layouts use white and soft lavender, readable measures and responsive stacked sections. Article CSS is scoped; footer/navigation typography is shared. Discovery and organizer sign-in use scoped 16–17px reading styles without changing their data or authentication workflows.

## Product accuracy

- FAQ copy lives in `components/site/FAQ.tsx`. Landing questions and the full help topics use the same source. Update it when registration, account or management behavior changes.
- Current copy covers direct/approval registration, optional guest entry, waiting-list offers, eligible cancellation/notes/transfer, guest registration claiming, tickets, check-in, Spaces, owner-only guest records, event visibility and editing conflicts.
- Search currently matches titles/descriptions. Do not advertise a format filter until the attendee UI provides one.
- Private events are link-only, not an invitation allowlist. Queued email is not confirmed delivery. Payments, password recovery, self-service account deletion/export and offline scanning are not advertised as available.
- Recheck current handlers and UI, rather than deriving behavior from older architecture plans or roadmap entries. Cookie duration and scope reflect the mounted session implementation; the reserved admin cookie is not described as an active customer cookie.

## Support and policy details

The user confirmed **SUPPORT@hackota.com** on 20 October 2026. It is the public support fallback in `lib/site.ts` and the example environment value. A valid `NEXT_PUBLIC_SUPPORT_EMAIL` can configure an environment-specific public address. The contact flow is an actual mailto link, not a simulated form submission or delivery acknowledgement.

The policy text documents current behavior and responsibilities. It does not certify legal compliance, guarantee email delivery, invent a company identity/address or promise a retention deadline. The legal operator name/address, a defined retention schedule and a platform minimum-age policy were not supplied and still need business/legal confirmation before representing the notices as a complete jurisdiction-specific legal policy. Do not infer the legal operator from the support email domain. Published policy text is dated 20 October 2026.

Official browser references linked from Cookies: [Chrome cookie controls](https://support.google.com/chrome/answer/95647) and [Firefox site-data controls](https://support.mozilla.org/en-US/kb/clear-cookies-and-site-data-firefox). Any future jurisdiction-specific legal language should be checked against current primary legal sources and the operator's actual practices.
