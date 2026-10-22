# Public media domain

Target public URL: `https://cdn.hackways.in/media-uploads/<blob-key>`.

## Current state (2026-10-22)

- Cloudflare is authoritative for `hackways.in`.
- Added an explicit proxied CNAME from `cdn.hackways.in` to
  `hackwaysstorage.blob.core.windows.net`; the existing VM wildcard is unchanged.
- Storage account `hackwaysstorage` is in `CRIN-MDB-RG-PRO`. HTTPS is required.
- `media-uploads` permits anonymous blob reads, but not container listing.
- `backups` is private. Keep its access unchanged.
- The Go signer and banner-completion handler already support `AZURE_CDN_DOMAIN`.
  Compose now forwards it, with an empty default preserving the Azure origin.
- Azure Cloud Connector and the scoped public-read block rule are active. Existing
  connector/custom-rule entrypoints were absent; no prior rules were replaced.
- HTTPS GET/HEAD/range checks pass for an existing JPEG; SHA256 matches Azure and
  repeated reads are cache hits. CDN root/listing/backups/write probes return 403.
  Azure CORS returns the matching `https://hackways.in` origin on image GET.
- Activation is still pending: default Cloudflare caching stores 404 responses.
  The supplied token lacks Cache Rules edit access (API returns 403), so
  `media-cache-rule.json` has been prepared but not applied. Global-backend runtime
  still uses Azure public URLs; no containers have been replaced.
- Before/after edge configuration and delivery evidence are stored outside Git in
  the user-only `Documents/HackwaysCDN` directory. The API token is not stored there.

## Edge configuration

Use Cloudflare's existing plan and Cloud Connector for Azure Blob Storage. Do not
provision Azure Front Door/CDN or add a VM media proxy. Cloud Connector is currently
available on the free plan; Azure storage/egress charges still apply.

1. Read and save the existing `cdn` DNS record, Cloud Connector rules, relevant
   cache/security rules and TLS configuration before changing anything. Confirm
   whether `cdn` is an explicit record or inherited from a wildcard.
2. Configure the rule in `media-connector-rule.json` using the Microsoft Azure
   provider. Preserve the path including `/media-uploads/`; do not rewrite it.
   This rule handles public GET/HEAD requests without query strings. Upload SAS
   URLs continue using Azure directly.
3. Update only the exact `cdn.hackways.in` record to `media-dns.json`, proxied.
   Leave wildcard/other application DNS records unchanged. Cloud Connector sets
   the origin host to `hackwaysstorage.blob.core.windows.net`; Azure custom-domain
   registration is unnecessary for this route. Require verified HTTPS to Azure
   and a valid Cloudflare edge certificate for `cdn.hackways.in`.
4. Use `media-block-rule.json` to scope a custom block rule to `cdn.hackways.in` for requests outside the same
   allowed path/method/query conditions above. Confirm rule quota before adding;
   never replace an unrelated rule to make space. This keeps container listing,
   backups paths, API routes, write methods and signed queries off the CDN.
5. Apply `media-cache-rule.json` to the cache rules entrypoint after reading and
   preserving any existing cache rules. It honors origin Cache-Control, respects
   browser TTL and applies `no-store` (-1) to status codes 400-599. Inspect existing
   cache overrides; never cache signed queries or private data. The checked JPEGs
   have no explicit origin Cache-Control, so successful images otherwise use
   normal Cloudflare defaults. Do not impose immutable caching on avatars, since
   their write SAS currently allows replacement. Check missing-file behavior with
   a fresh unused key after applying this rule, and account for already cached 404s.

Cloud Connector API updates replace the complete rule list. GET
`/zones/{zone_id}/cloud_connector/rules` first, merge this single rule with every
existing rule (preserving their IDs), then PUT the complete array to that endpoint.
Do **not** PUT this object alone or replace the list with a one-rule array. Use
zone-scoped DNS Edit and Cloud Connector Write permissions; an edge block-rule
change additionally needs the appropriate zone WAF edit permission. Cache rules
need zone Cache Rules Edit (Cache Settings Write) permission. Keep tokens
outside the repository and out of logs.

## Activation checks

Before setting the application environment variable:

- Use an existing public blob. Compare HTTPS GET status, MIME type, length and
  SHA256 through Azure and CDN. Check certificate validation, HEAD and byte range
  behavior; repeat the GET and inspect `CF-Cache-Status`/`Age` for eligible images.
- Confirm missing objects return a genuine 404 without cached failure, and the
  root, listing queries, backups path and write methods are blocked at the edge.
- Confirm private backups remain private at their Azure origin.
- Use an identifiable curl/browser client for delivery checks. Default Python
  urllib requests currently receive Cloudflare Browser Integrity Check error 1010;
  this zone-wide setting was not changed or bypassed with a skip rule.
- Inspect Blob CORS if assets are fetched by script/canvas/fonts; permit only the
  required app origins. Do not broaden upload CORS or add credentials to CDN reads.
- Set `AZURE_CDN_DOMAIN=https://cdn.hackways.in` in the managed global-backend
  runtime environment and recreate only that service using its approved current
  immutable image and the existing release manifest. No rebuild is needed for
  the existing URL behavior. Do not release unrelated local work.
- Exercise the normal authenticated upload intent/completion path: write SAS
  host must remain Azure, public URL must use CDN, and the confirmed stored banner
  must load. These mutation checks require an approved event/account.

New upload URLs use CDN after runtime activation. Existing persisted Azure URLs
continue to work; this configuration does not rewrite historical database rows.
Migrate old URLs separately only after CDN verification, with an exact origin and
container match and an audited reversible migration. Private URLs must never be
rewritten. Application JS/CSS under `/_next` remains tied to each independent
application release; do not point four apps at one shared unversioned asset prefix.
Public media/static assets stored in `media-uploads` can use this domain.

## Rollback

Remove the global-backend CDN environment value first to restore Azure URLs for
new uploads, preserving all approved images. Keep the CDN serving any URLs already
stored or sent by email; deleting the route would break those links. Restore edge
configuration only after accounting for such references. Preserve saved DNS/rule
backups outside Git. Never reset blob permissions or delete files during rollback.

## References

- [Cloudflare Azure Blob example](https://developers.cloudflare.com/rules/cloud-connector/examples/serve-static-assets-from-azure/)
- [Cloud Connector availability and behavior](https://developers.cloudflare.com/rules/cloud-connector/)
- [Cloud Connector API and replacement semantics](https://developers.cloudflare.com/rules/cloud-connector/create-api/)
- [Azure custom-domain HTTPS requirements](https://learn.microsoft.com/en-us/azure/storage/blobs/storage-custom-domain-name)
