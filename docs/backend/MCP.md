# Hackways MCP implementation

## Requested scope and service boundaries

The operator requested a remote MCP at `https://mcp.hackways.in/mcp`, a dedicated
consent page, and attendee, organizer and administrator workflows with many focused
tools. Preserve the four Next.js apps and two Go engines: global-backend hosts MCP
transport and OAuth; tools invoke the existing authenticated Go HTTP APIs. Core
continues to own event/registration/arrival business logic. Next renders consent,
action approvals and connection management only. No TypeScript business backend,
new database, queue or runtime service is introduced.

Use the official Go MCP SDK. Version 1.8.0 requires Go 1.25 (verified in its module
manifest); update only global-backend's build baseline and let CI select each Go
module's declared version. Retain existing hard limits and bounded pools. Verify
runtime resource behavior; SDK adoption is not a throughput claim.

## Protocol and identity contract

- Stateless Streamable HTTP with JSON responses, explicit input/output schemas,
  honest tool annotations and per-tool OAuth security schemes. Public discovery is
  anonymous; private tools always validate credentials. No generic SQL/raw-API tool.
- Global OAuth 2.1 authorization code plus mandatory PKCE S256; resource/issuer
  binding, exact registered redirects, issuer identification, discovery, client
  registration/CIMD, single-use codes, short opaque hashed access tokens and rotating
  hashed refresh credentials. No session cookie or password enters a tool argument.
- Consent pages on the existing main/admin origins retain host-only cookies.
  Administrator permissions require a real dedicated admin session and daily OTP.
  OAuth grants bind to the actual session; current session/grant/role/revocation
  checks stay authoritative. Normal account authentication is not rewritten.
- Existing HTTP APIs accept these distinct OAuth credentials only for explicitly
  mapped routes/scopes. Unknown routes default to denied. Existing ownership,
  capacity locks, revisions and admin grants/OTP continue to apply.
- Every OAuth write additionally requires a short-lived browser-approved action
  intent bound to its grant, API service, method, canonical path and exact payload.
  A bearer token alone cannot bypass confirmation by calling the HTTP API directly.
  Stable intent keys feed existing domain idempotency. Store only real operation
  outcomes; an ambiguous failure must direct the user to inspect saved state.

## User journeys

A connection shows verified/unverified client identity, requested permissions,
account and explicit Allow/Cancel actions. Admin consent uses the separate admin
origin and existing verification. The account page lists connections and supports
revocation. Expiry, cancelled consent, denied scope, changed roles, revoked sessions,
refresh reuse and database outages have explicit responses.

Tools cover public discovery, own profile/registrations/pass links, organizer
creation/editing/questions/guests/invitations/check-in/Spaces and admin records,
analytics, campaigns/custom messages, promotion scheduling and feature rollouts.
Sensitive pass QR credentials and storage SAS URLs are not returned to the model;
provide the existing authenticated browser pages instead. Never send a real
campaign/message or change real records while verifying the implementation.

## Deployment and verification

Require numbered SQLc-backed migrations and existing mandatory isolated CI suites.
Verify PKCE, code/refresh reuse, audience/scope/redirect/origin checks, anonymous and
cross-account denial, admin OTP/grant revocation, approval binding and real tool
results/errors. Run protocol initialization/list/call checks and the MCP Inspector
where available. No invented fixtures in application paths or claimed browser/
OpenAI review success. Build images in CI. Configure DNS, TLS, streaming proxy,
metadata routing, rate limits and private backend ports before public activation.
Do not claim Directory publication without actual submission/approval.

## Official guidance consulted

- https://developers.openai.com/plugins
- https://developers.openai.com/plugins/build/mcp-server
- https://developers.openai.com/plugins/build/auth
- https://developers.openai.com/plugins/plugin-guidelines
- https://developers.openai.com/plugins/deploy/app-review

Implementation uses 106 focused tools. Public landing: `https://hackways.in/mcp`.
Connection reference: `https://hackways.in/mcp/setup`, covering CIMD, DCR and manual
public-client registration, exact endpoints, connection-specific callbacks, all
eight scopes, mandatory S256 PKCE and explicitly disabled OIDC. The page uses
deployed Go metadata as its configuration authority; it creates no credentials.
Transport/resource: `https://mcp.hackways.in/mcp`; issuer: `https://mcp.hackways.in`.
Dedicated account/admin consent pages are `/connect/authorize`, action review is
`/connect/action`, and revocation is `/connect/connections`. Mixed permission
requests start on the account origin with admin scopes omitted; a separate link
opens dedicated admin consent. Only actually approved scopes reach the grant.

The SDK owns Streamable HTTP, input/output validation and tool dispatch. A bounded
JSON adapter adds root `securitySchemes` to `tools/list` because Go SDK Tool has
only `_meta` extension support; both root and compatibility mirror are verified.
Read-only previews remain reads; all actual writes require an approved exact
request and private execution lease. Unknown API routes fail closed. Lost execution
responses are uncertain and never automatically repeated. Repeated completed
approvals return the persisted actual result without invoking the API again.

Runtime switch `MCP_ENABLED` defaults false on both Go services. Apply migration 24
before activation; global alone advances to Go 1.25 for the official SDK. Keep
existing pools, session/signing identity and CPU/memory limits. CI supplies images;
no frontend build runs on the VM. Four concurrent tool operations, 128 KiB input,
512 KiB upstream responses, 1 MiB catalog adaptation, 12-second upstream timeout,
10 pending actions per grant and 50 active connections per account bound work.
Client registry holds at most 2,000 entries, authorization requests at most 10,000,
and active access credentials at most four per grant. OAuth registry/auth creation
is serialized transactionally; unused clients expire after seven days and bounded
housekeeping removes expired requests/actions/access/grants. DCR additionally has
one registration/minute/IP with a small Nginx burst. These are admission limits,
not measured throughput claims.

The MCP certificate was issued and externally verified with the system trust store
on 23 October 2026. The temporary HTTPS preparation response remains until the
verified runtime release. The certificate uses certbot's existing webroot renewal;
the production vhost retains its HTTP challenge path. Application activation,
Linux race CI, image publication and actual browser/OpenAI client review remain
pending. No Directory submission or approval is claimed.

Previous release: analytics/check-in source
`1e898be5e702b5a48e12135f333ba68117839b26`, schema 23, all six services live;
CI 39901702575 passed and PR 46 merged. CDN and Apple Wallet remain deferred.

### ChatGPT client metadata negotiation

The production metadata document was fetched successfully from the VM on
23 October 2026. Its plural supported-method array contains none and
private_key_jwt while its legacy singular preference is private_key_jwt.
Use the documented method intersection and select none, which is the only method
this authorization server advertises. PKCE remains mandatory; client assertions
and secrets are rejected rather than silently treated as authenticated clients.
The signed assertion method is not advertised or claimed. This compatibility
correction is independently built as a global-backend-only candidate so public
footer publication is not delayed.

The designated profile tool returns the standard top-level persisted account ID, name and email schema. Authentication challenges include both error and error_description as required for the OpenAI linking UI. Other tools retain their actual-operation response envelopes.
