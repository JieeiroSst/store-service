# Contract test checklist — Go backend vs. `linkforty_flutter` SDK

This backend is a from-scratch Go rewrite of `@linkforty/core` (TypeScript).
The Flutter SDK only ever talks REST/JSON — it does not know or care that
the backend changed language — so the pass/fail bar is: **every field name,
shape, and status code the SDK sends/reads must match byte-for-byte.**

## 1. Setup

```bash
# Postgres + (optional) Redis
docker run -d --name lf-pg -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=linkforty postgres:16
docker run -d --name lf-redis -p 6379:6379 redis:7

# .env (see .env.example)
DATABASE_URL=postgresql://postgres:password@localhost:5432/linkforty
REDIS_URL=redis://localhost:6379
PORT=3000
APP_ENV=development
CORS_ORIGIN=*
TRUST_PROXY=
IOS_TEAM_ID=ABC123XYZ
IOS_BUNDLE_ID=com.example.app
ANDROID_PACKAGE_NAME=com.example.app
ANDROID_SHA256_FINGERPRINTS=AA:BB:...
SHORTLINK_DOMAIN=http://localhost:3000

cd shortlink-service
go run ./cmd/server
# migrations apply automatically on startup (golang-migrate, embedded SQL);
# `migrate -path migrations -database "$DATABASE_URL" up` also works standalone
```

Sanity check before touching the SDK at all:

```bash
curl -s localhost:3000/health              # {"status":"ok","uptime":N}
curl -s localhost:3000/health/ready        # {"status":"ok","checks":{"database":"ok",...}}
curl -s localhost:3000/api/sdk/v1/health   # {"status":"healthy","version":"v1","timestamp":"..."}
```

## 2. Point the Flutter SDK at this backend in debug mode

In the app using `linkforty_flutter`, set:

```dart
await LinkForty.initialize(LinkFortyConfig(
  baseUrl: 'http://<your-machine-ip>:3000', // https required unless localhost
  debug: true, // verbose SDK logging — prints every outgoing request/response
));
```

`debug: true` is the important bit here: it logs each request URL, body,
and the raw response the SDK parsed. Run the app through the flows below
and diff the logged request/response against what this Go server actually
received/returned (`go run ./cmd/server` also logs every request via the
zap production logger — cross-reference request IDs/timestamps).

Alternative capture method if the SDK's debug logs aren't detailed enough:
run `mitmproxy` or `proxyman` between the device/simulator and the Go
server, or temporarily point `baseUrl` at a `webhook.site`-style echo
first to see the exact raw payload before switching it to the real server.

## 3. Endpoint-by-endpoint checklist

For each row: trigger the SDK flow, capture the request the SDK actually
sent, and confirm the Go server's response has every field the SDK's
model deserializer expects (extra fields are harmless; **missing or
misnamed fields are not**).

| # | Flow | SDK call | Endpoint | What to verify |
|---|------|----------|----------|-----------------|
| 1 | First app launch | automatic on `initialize()` | `POST /api/sdk/v1/install` | Request has `userAgent`, `platform`, `platformVersion`, `deviceId`, `timezone`, `language`, `screenWidth/Height`. Response has `installId`, `attributed`, `confidenceScore`, `matchedFactors` (array, not null), `deepLinkData` (object, `{}` when organic). |
| 2 | Deferred deep link (install after a click) | automatic, same call as #1 | `POST /api/sdk/v1/install` | Click a real short link on the test device first (Safari/Chrome, not an in-app browser), uninstall/reinstall or clear app data, then launch. `attributed` should be `true`, `confidenceScore >= 70`, `matchedFactors` non-empty, `deepLinkData.shortCode` matches the clicked link. **This is the fingerprint-hash path — if this stays `false` every time, the fingerprint algorithm mismatched; see §4.** |
| 3 | Universal Link / App Link open (app installed) | SDK's URL handler → resolve | `GET /api/sdk/v1/resolve/:shortCode` or `/:templateSlug/:shortCode` | Tap a short link with the app installed. Response has `shortCode`, `linkId`, and whichever of `deepLinkPath/appScheme/iosUrl/androidUrl/webUrl/utmParameters/customParameters` the link actually has configured. `clickedAt` present. |
| 4 | Event tracking | `LinkForty.trackEvent('some_event', properties: {...})` | `POST /api/sdk/v1/event` | Request has `installId`, `eventName`, `eventData`. Response `{eventId, acknowledged: true}`. |
| 5 | Revenue tracking | `LinkForty.trackRevenue(amount: ..., currency: ...)` | `POST /api/sdk/v1/event` | `eventData` contains `revenue`/`amount` + `currency` fields per the SDK's convention — confirm the Go server stores/forwards `eventData` opaquely (it does; this port never inspects `eventData` contents, matching upstream). |
| 6 | Attribution lookup | `LinkForty.getAttributionData()` (if it hits network) or manual debug call | `GET /api/sdk/v1/attribution/:fingerprint` | 404 body `{"error": "..."}` when unknown; 200 with `installEvent`/`clickEvent`/`linkData` when known. |
| 7 | Programmatic link creation | `LinkForty.createLink(...)` (no templateId) | `POST /api/links` | Confirm `originalUrl`, `title`, `customCode`, `utmParameters`, etc. round-trip. **Response is a raw-row shape with duplicate snake_case + camelCase keys (see `internal/adapters/http/links_handler.go`'s doc comment) — if the SDK only reads `shortCode`/`id`/`original_url` this is fine; if it validates the full shape strictly, this needs tightening.** |
| 8 | Direct deep link / redirect | tap a short link in a regular (non-in-app) mobile browser | `GET /:shortCode` | 302 to the App/Play Store when app not installed, interstitial HTML when `app_scheme` is set, correct UTM params appended. |
| 9 | iOS Universal Links | app installed, tap link in Safari/Messages | `.well-known/apple-app-site-association` | `curl -s $BASE/.well-known/apple-app-site-association` returns `{"applinks":{"apps":[],"details":[{"appID":"<TEAM_ID>.<BUNDLE_ID>","paths":["*"]}]}}` with `Content-Type: application/json`, no `.json` extension in the path. |
| 10 | Android App Links | app installed, tap link in Chrome | `.well-known/assetlinks.json` | Matches `ANDROID_PACKAGE_NAME` + `ANDROID_SHA256_FINGERPRINTS`. |

## 4. Fingerprint/attribution deep check (highest-risk area)

If step 2 above never attributes:

1. Confirm `TRUST_PROXY` matches your topology. If the Go server sits
   behind any reverse proxy / load balancer / ngrok tunnel and `TRUST_PROXY`
   is unset, every request's IP resolves to the proxy's IP, not the
   device's — every click and every install will look like they share one
   IP, which can accidentally *help* IP-only matching but breaks
   `isAttributableIp()`'s assumption in subtle ways. Set `TRUST_PROXY=1`
   for a single reverse-hop deployment (nginx/Caddy in front) and verify
   with:
   ```bash
   curl -s -H "X-Forwarded-For: 1.2.3.4" localhost:3000/api/sdk/v1/health
   # then check the click_events.ip_address column after a real click through the proxy
   ```
2. Query the DB directly to compare the two fingerprint hashes:
   ```sql
   select fingerprint_hash, ip_address, user_agent, timezone, language, platform, platform_version
   from device_fingerprints order by created_at desc limit 5;
   select fingerprint_hash, ip_address, user_agent, timezone, language, platform, platform_version, confidence_score, matched_factors
   from install_events order by installed_at desc limit 5;
   ```
   The two hashes will legitimately differ (click-side IP/UA differs
   slightly from install-side by design — that's why matching is
   probabilistic, not exact-hash lookup). What to check instead:
   `matched_factors` should include `ip` when both sides are on the same
   network, and the confidence score should be >= 70.
3. If confidence never reaches 70: the SDK's `attributionWindowHours`
   default (168) must be within the link's own `attribution_window_hours`,
   and the click must be within `2160` hours (90 days) — both enforced in
   `internal/app/attribution.go`.

## 5. Known, deliberate deviations from upstream (re-verify these are OK)

1. **Validation-error status codes are 400, not upstream's accidental 500.**
   In the TS source, zod's `.parse()` on `POST /api/links` and
   `POST /api/sdk/v1/event` runs *before* the route's try/catch, so a
   malformed request actually gets HTTP 500 with a Fastify-default error
   body upstream — almost certainly a latent bug, not a contract. This
   port returns conventional 400s instead. If the Flutter SDK's error
   handling specifically branches on 500 vs 400 for a malformed request,
   this would need to change back to match upstream exactly. Low risk:
   a correctly-implemented SDK should never send a malformed request to
   these two endpoints.
2. **GeoIP data source differs** (MaxMind GeoLite2 `.mmdb` via
   `GEOIP_DB_PATH`, vs. upstream's bundled `geoip-lite` dataset). Field
   *names* match (`countryCode`, `countryName`, `region`, `city`,
   `latitude`, `longitude`, `timezone`); per-IP *values* may differ.
   Doesn't affect the fingerprint hash or attribution scoring (neither
   reads geo data). Unset `GEOIP_DB_PATH` → all geo fields null, same as
   upstream's "IP not found" case.
3. **`isbot` bot-detection database** is approximated with a hand-written
   pattern list (`internal/domain/bot_detection.go`) — no Go equivalent of
   the curated `isbot` npm package exists. Affects only the `is_bot`
   analytics flag, not any SDK-facing field.
4. **`ua-parser-js` OS/browser string output** is approximated
   (`internal/domain/user_agent.go`). Confirmed from upstream's own
   `calculateConfidenceScore()` that `platformVersion` is stored but never
   scored in attribution matching — this only affects cosmetic analytics
   columns (`click_events.platform`), not attribution quality or any
   SDK-facing response field.
5. **OG preview route** (`/:shortCode/preview` + the social-scraper
   auto-detect hook) and **`/api/debug/*`** (including the live click
   WebSocket) are **not ported in this pass**. Both are opt-in upstream
   too — `createServer()` never auto-registers `previewRoutes()` or
   `debugRoutes()` — so this doesn't regress default behavior, but if your
   deployment relied on either, they need to be added.
6. **`POST /api/links` response shape**: reproduces upstream's literal
   object-spread quirk of including both snake_case raw columns and
   camelCase overrides as separate JSON keys, *except* `click_count`
   (upstream: raw Postgres `COUNT(*)` as a numeric string, alongside
   `clickCount` as a real number) — this port emits `click_count` as a
   number in both places rather than reproducing the string/number split.

## 6. Data model spot-checks

```sql
-- after running through flows 1-4 above, confirm no NULL-constraint
-- surprises and that JSONB columns aren't storing the literal string "null"
select * from links order by created_at desc limit 3;
select * from click_events order by clicked_at desc limit 3;
select * from install_events order by installed_at desc limit 3;
select * from in_app_events order by event_timestamp desc limit 3;
select * from device_fingerprints order by created_at desc limit 3;
```

## 7. Webhooks (if configured)

```bash
curl -X POST localhost:3000/api/webhooks -H 'Content-Type: application/json' -d '{
  "name": "test", "url": "https://webhook.site/<your-id>",
  "events": ["click_event", "install_event", "conversion_event", "sdk_event"]
}'
curl -X POST localhost:3000/api/webhooks/<id>/test
```
Confirm the delivered payload has header `X-LinkForty-Signature: sha256=<hex>`
and that `hmac_sha256(body, webhook.secret) == <hex>` (see
`internal/domain/webhook.go`'s `GenerateWebhookSignature`, identical
algorithm to upstream's `generateWebhookSignature()`).
