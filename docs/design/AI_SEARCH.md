# Public search and AI discovery

This implementation applies the user-requested `ai-seo` skill to Hackways' public
information. The site remains an event/community platform, with private account,
registration and application journeys served by the existing Go APIs and apps.

## Implemented

- The homepage emits server HTML JSON-LD for the factual Organization and WebSite
  identity. Its name, purpose, public URL and brand image agree with the About
  page. There are no invented addresses, awards, reviews, ratings, social
  accounts, salaries or paid-employment claims. The script is serialized with
  `<` escaped and includes no user-generated or applicant data.
- About introduces the platform in a direct, self-contained sentence. Existing
  substantive organizer, help and policy pages remain the sources for details.
- `/llms.txt` is a small static public directory. It links to existing event,
  community, program, help, policy and MCP pages rather than duplicating changing
  event dates or application status. Current availability/requirements remain on
  the actual public pages. No `llms-full.txt` is published because a static file
  would not accurately represent every dynamic record on the platform.
- The earlier audit fixes add public directory sitemap URLs, public canonical
  URLs and specific directory descriptions, retaining distinct community
  pagination. Search and private community views are noindex. See the
  [dated evidence and limits](../planning/SEO_AUDIT_2026-10-26.md).
- Circular brand icons are explicit in all app roots. See
  [brand assets](BRAND_ICONS.md).

## Crawler policy and evidence

The existing production robots file allows `*`; no discovery or training policy
has been changed. On 2026-10-26 anonymous `/about` requests using the
OAI-SearchBot, PerplexityBot and Claude-SearchBot user-agent headers returned HTTP
200 with the public content and no `cf-mitigated` challenge header. This checks
header-based behavior from the audit connection only. It does not establish that
requests from verified crawler IP ranges are allowed or that a crawler has indexed
the site. No WAF/auth/rate-limit exception was added.

Search crawling, user-triggered retrieval and model training are separate
controls. OAI-SearchBot and GPTBot should not be treated as equivalent; see the
[official crawler documentation](https://developers.openai.com/api/docs/bots).
Google states that its AI search features need no special AI files or markup;
normal technical SEO and helpful content remain the foundation. `llms.txt` is an
optional discovery aid, not a ranking requirement or citation guarantee. See
[Google's AI features guidance](https://developers.google.com/search/docs/appearance/ai-features).

## Measurement still required

No Search Console, AI citation dashboard or analytics account was connected for
this work. No measured ranking, traffic, recommendation rate or AI visibility lift
is claimed. The earlier PageSpeed request returned429; field Core Web Vitals are
still unmeasured. JSON parsing and server-output checks do not replace Google's
rendered Rich Results Test or a browser accessibility review.

Use this small prompt set when establishing a real baseline:

| Prompt family | First-party source to check | Status |
| --- | --- | --- |
| What is Hackways? | `/about`, homepage identity | Not measured |
| How do I find events on Hackways? | `/discover`, `/help/finding` | Not measured |
| How do event registrations work? | `/help/joining`, each event's page | Not measured |
| How do I host an event on Hackways? | `/organizers`, `/help/hosting` | Not measured |
| What are Hackways Spaces? | `/spaces`, `/about` | Not measured |
| How do Hackways Ambassadors applications work? | `/ambassadors`, current program page | Not measured |
| How do I connect Hackways to an AI workspace? | `/mcp`, `/mcp/setup` | Not measured |

For each engine record the exact prompt, date, model/surface, locale, cited URL,
factual answer and sample size. Run each prompt several times before interpreting
a rate. Fix incorrect first-party facts at their owning public page; do not
manufacture independent endorsements or inject instructions into content.

## Operational boundary

This is public content/metadata and static asset work. No schema migration, Go
business logic, form input, authentication, submission, email or application
decision contract changes are part of it. Production deployment still requires
the existing off-VM CI and scoped runtime release; do not restart live applicant
services merely to validate local SEO output.
