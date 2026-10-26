# Google Search Console setup for Hackways

## 1. Create the property

Open [Google Search Console](https://search.google.com/search-console) using the
Google account that should own Hackways' search data. Choose **Add property**,
select **Domain**, and enter `hackways.in` without `https://` or a page path.
This covers the domain's protocols and subdomains.

## 2. Add Google's verification record

Copy the unique TXT value shown by Google. In Cloudflare, open the `hackways.in`
zone, then **DNS > Records > Add record**:

| Field | Value |
| --- | --- |
| Type | TXT |
| Name | `@` |
| Content | The exact `google-site-verification=...` value Google supplied |
| TTL | Auto |

Save it as an additional record, leaving the current mail verification, SPF and
other TXT records intact. Return to Search Console and click **Verify**. If Google
has not seen the record yet, allow DNS propagation and try again. Keep this TXT
record after verification so ownership can continue to be checked.

[Google's ownership verification instructions](https://support.google.com/webmasters/answer/9008080)

## 3. Submit the sitemap

Open **Sitemaps** in the verified property. Submit:

`https://hackways.in/sitemap.xml`

Also submit the automatically refreshed public catalog sitemap:

`https://hackways.in/catalog-sitemap.xml`

Confirm that Google can fetch it and review the reported status. This sitemap
lists public pages and directories; private applications and account pages are
not indexing targets.

[Google's sitemap report instructions](https://support.google.com/webmasters/answer/7451001)

## 4. Inspect the important pages

Use the URL Inspection bar for these full URLs:

- `https://hackways.in/`
- `https://hackways.in/discover`
- `https://hackways.in/ambassadors`
- `https://hackways.in/programs`
- `https://hackways.in/programs/ambassadors`

For each, run **Test live URL**. If the page is available and indexable, use
**Request indexing** where offered. After Google has crawled a page, compare its
user-declared canonical with Google's selected canonical. A live test cannot
predict which canonical Google will select.

[Google's URL Inspection instructions](https://support.google.com/webmasters/answer/9012289)

## 5. Review the reports

Check **Page indexing**, **Performance > Search results**, **Core Web Vitals**,
**Manual actions** and **Security issues**. Private/login/manage/application URLs
being excluded by noindex is expected. Public pages unexpectedly blocked,
returning errors or assigned an incorrect canonical deserve investigation.

The domain property includes the admin subdomain; its existing noindex policy
should remain. Search Console setup does not make private pages public.

## 6. Establish a baseline

Record public-page impressions, clicks, CTR, queries and chosen canonicals once
data appears. Setup and indexing requests do not guarantee indexing, a search
position, immediate favicon refresh or AI citations. The site's favicon and
public metadata can be fetched independently while Google processes updates.

No verification token is built into source code for this DNS method. No Google
Analytics or Tag Manager installation is required to verify the Domain property.

## If Discover is reported as soft 404

A successful fetch and allowed indexing do not guarantee Google considers the page
useful. The live event API returned zero upcoming public events when this was
reported. Discover now has substantive server-rendered context and useful links
in addition to its honest live catalog. After the release, inspect
`https://hackways.in/discover`, run Test live URL and request indexing where
offered. Monitor the indexed report after Google recrawls; a local content change
cannot confirm that Google has cleared its classification.

Google discovers additions through the two sitemaps and internal links. This is
automatic URL discovery, not guaranteed automatic inclusion in search results.
The [Indexing API](https://developers.google.com/search/apis/indexing-api/v3/quickstart)
is restricted to eligible job postings and livestreaming video pages; ordinary
event, community and marketing pages should use sitemaps and Search Console.
