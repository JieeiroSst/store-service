# bot-service

A multi-channel chat bot **and** social content scheduler: it replies to inbound messages on Telegram/Twitter(X)/Facebook, and separately publishes scheduled or recurring content posts to those same channels.

```
Telegram long-poll  ─┐                                   ┌─ Telegram (reply / broadcast)
X (Twitter) mentions ─┼─ BotUsecase.HandleMessage ─ reply ┼─ Twitter  (reply / new tweet)
Facebook webhook    ─┘                                   └─ Facebook (Send API / Page feed)

HTTP /v1/posts (create/list/get/publish/cancel)
        │
        ▼
ContentUsecase ── scheduler (robfig/cron) ── ChannelPublisher (per channel) ── social API
        │
        ▼
   Postgres (posts table)
```

## Architecture

Hexagonal (ports & adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
config/                         env-var config, with a Consul KV override (see below)
internal/
  domain/
    model/                      Channel, IncomingMessage/OutgoingMessage, Post, ...
    port/
      driving.go                BotUsecase, ContentUsecase (primary ports)
      driven.go                 MessageSender/ChannelSender, ChannelPublisher, PostRepository (secondary ports)
  application/
    bot.go                      BotUsecase impl: echoes inbound messages back on the same channel
    content.go                  ContentUsecase impl: create/publish/cancel posts, one-off + recurring
  adapter/
    primary/                    drives the application
      telegram/                 long-polling updates -> BotUsecase
      twitter/                  interval polling of mentions -> BotUsecase
      http/                     health check, content management API, Facebook webhook
      scheduler/                cron-driven trigger -> ContentUsecase.PublishDuePosts
    secondary/                  driven by the application
      telegram/, twitter/, facebook/   ChannelSender (replies) + ChannelPublisher (content posts) per platform
      router/                   fans MessageSender out to the right ChannelSender by model.Channel
      repository/               Postgres-backed PostRepository (gorm)
  infrastructure/                fx module wiring, config/logger bootstrap, lifecycle-managed servers
```

Two independent things share the domain's `Channel` concept but are otherwise separate flows:

- **Replying** (`BotUsecase` / `ChannelSender`): triggered by an inbound message (Telegram update, Twitter mention, Facebook Messenger event), always replies in the same conversation it came from.
- **Content publishing** (`ContentUsecase` / `ChannelPublisher`): triggered by the HTTP API or the scheduler, posts *new* content to a channel's broadcast target (a Telegram chat/channel, a new tweet, a Facebook Page's feed) — independent of any inbound conversation.

## Setup

```bash
cp .env.example .env   # fill in TELEGRAM_BOT_TOKEN at minimum
for f in migrations/*.sql; do psql "$DATABASE_URL" -f "$f"; done
go run .
```

Schema changes live in `migrations/` as plain numbered SQL files (`001_...`, `002_...`, applied in order) — there's no migration runner or `AutoMigrate` wired up, so new migrations are written by hand and applied by hand, same as the rest of this repo's services.

### Configuration

Config resolves from **Consul KV** first (bootstrapped via `.env`'s `HostConsul`/`KeyConsul`/`ServiceConsul`, JSON shape in `consul.json`), falling back to plain env vars if Consul is unset or unreachable — see `internal/infrastructure/config_provider.go`. `.env.example` lists every fallback env var; the ones that matter to get started:

| Env var | Default | Description |
|---|---|---|
| `PORT_HTTP_SERVER` | `8080` | Health check + content API + Facebook webhook |
| `POSTGRES_*` | `localhost:5432/bot_service` | Post storage |
| `SCHEDULER_POLL_INTERVAL` | `30s` | How often the scheduler checks the DB for due posts. This is **not** a schedule - it's just the check frequency; every post's actual run time lives on the post itself (`scheduled_at`/`cron_expr`, see below), never in config |
| `TELEGRAM_BOT_TOKEN` | — | Required to enable Telegram at all |
| `TELEGRAM_BROADCAST_CHAT_ID` | — | Chat/channel ID content posts are published to |
| `TWITTER_ENABLED` | `false` | Twitter reply-mentions polling + content publishing |
| `TWITTER_USER_ID`, `*_TOKEN*` | — | X API v2 credentials (OAuth 1.0a for posting, bearer for reading mentions) |
| `FACEBOOK_ENABLED` | `false` | Facebook Messenger webhook + Page content publishing |
| `FACEBOOK_PAGE_ID`, `FACEBOOK_PAGE_ACCESS_TOKEN`, `FACEBOOK_VERIFY_TOKEN`, `FACEBOOK_APP_SECRET` | — | Graph API credentials |

## API

### Health

```
GET /health
```
```json
{"status": "ok"}
```

### Content management (`/v1/posts`)

This is the single source of truth every other part of the system (scheduler, publishers) reads from — every field below is a real column/JSON key on `model.Post`, not aspirational.

A **Post** has:
- **Content**: `title` (internal note/label, never sent to any channel), `text` (the caption), `hashtags` (`[]string`).
- **Media**: `media` — a list of `{url, type, thumbnail}` (`type` is `image`/`video`; `thumbnail` only matters for video). `media_kind` classifies the post as a whole (`text_only`/`single_image`/`multi_image`/`video`/`reel`) — auto-derived from `media` if omitted, except `reel`, which is a content-intent distinction (a short vertical video meant for Reels/TikTok) no file alone carries, so it's never guessed. `url` is expected to already be a hosted link (Drive/S3/Cloudinary/your own CDN/...) — this service stores references, not file bytes. Only the **first** media item is ever published (Twitter/Facebook multi-image galleries aren't implemented).
- **Platforms & schedule**: `channels` (one or more of `telegram`/`twitter`/`facebook`), and either `scheduled_at` (one-off) or `cron_expr` (recurring) — see below.
- **Post-publish tracking**: `results`, one `{channel, external_id, published_url, error, published_at}` entry per targeted channel (see below).
- **Admin**: `campaign` (free-form grouping/filter tag), `created_by`, `approved_by`, `status_changed_by` (see "Actor tracking").

Content is `text`, `media`, or both — the only rejected case is neither (`400 post must have text or media`). A text-only post sends a plain message; media-only sends the image with no caption; both sends the image captioned with `text`.

It targets one or more `channels` and runs either:
- **one-off** — publishes once at `scheduled_at` (omit it to publish at the next scheduler tick), or
- **recurring** — set `cron_expr` (a standard 5-field cron expression, or a `robfig/cron` descriptor like `@every 1h`, `@daily`); `scheduled_at` is ignored. It stays `active` and fires repeatedly until cancelled. `max_runs_per_day`/`max_runs_per_month` (0 = unlimited) cap how many of those firings actually publish — a schedule can fire more often than its cap allows, and the excess firings are skipped, not queued.

**Only `scheduled`/`active` posts are ever auto-published, and only via the cron scheduler.** `PublishDuePosts` fetches strictly `status = scheduled` (one-off, `DueForPublish`) or `status = active` (recurring, `ActiveRecurring`), and re-checks that status in Go before dispatching — a draft, a pending-review post, or anything else is never touched by the automated path. See "Preventing duplicate posts" below for how it also avoids double-sending an individual post.

#### Two ways to create one

**Draft first, schedule later** — save content as a `draft` with `save_as_draft: true`; `channels`/`scheduled_at`/`cron_expr` aren't required yet:
```json
{
  "save_as_draft": true,
  "title": "August flash sale",
  "text": "Flash sale starts now",
  "media": [{"url": "https://cdn.example.com/sale.jpg", "type": "image"}]
}
```
Then, once it's ready: fill in the rest with `PATCH /v1/posts/:id`, `POST /v1/posts/:id/submit` to move it to `pending_review`, and an editor `POST /v1/posts/:id/approve`s (→ `scheduled`/`active`) or `POST /v1/posts/:id/reject`s (→ back to `draft`, with a reason) it. See the workflow diagram below.

**Direct schedule** — omit `save_as_draft` (or set it `false`): `channels` is required and the post goes straight to `scheduled`/`active`, skipping review entirely. This is the quick path for one-off internal use.

```
draft ──submit──▶ pending_review ──approve──▶ scheduled / active ──▶ published / failed
  ▲                     │
  └───────reject────────┘
```
`cancel` works from any of `draft`, `pending_review`, `scheduled`, `active` → `cancelled` (a dead end).

#### Create a post

```
POST /v1/posts
Content-Type: application/json
```

One-off, text only, publish ASAP (direct schedule, no review):
```json
{
  "text": "New menu is live!",
  "channels": ["telegram", "facebook"]
}
```

One-off, scheduled, image with a caption:
```json
{
  "text": "Flash sale starts now",
  "media": [{"url": "https://cdn.example.com/sale.jpg", "type": "image"}],
  "channels": ["telegram"],
  "scheduled_at": "2026-08-10T09:00:00Z"
}
```

One-off, image only, no caption:
```json
{
  "media": [{"url": "https://cdn.example.com/sale.jpg", "type": "image"}],
  "channels": ["telegram"]
}
```

Recurring, daily at 9am, capped at one run a day, tagged with a campaign and hashtags:
```json
{
  "text": "Good morning! Today's special: ...",
  "hashtags": ["#promo", "#daily"],
  "channels": ["telegram", "twitter"],
  "cron_expr": "0 9 * * *",
  "timezone": "Asia/Ho_Chi_Minh",
  "max_runs_per_day": 1,
  "campaign": "morning-promo-q3",
  "created_by": "alice"
}
```

Response (`201 Created`, shape is `model.Post` — see below; zero-value fields like `max_runs_per_month`, `runs_today`, `last_run_at` carry `omitempty` and are absent until they matter):
```json
{
  "id": "3f9c8f1e-...",
  "text": "Good morning! Today's special: ...",
  "hashtags": ["#promo", "#daily"],
  "media": [],
  "media_kind": "text_only",
  "channels": ["telegram", "twitter"],
  "scheduled_at": "0001-01-01T00:00:00Z",
  "timezone": "Asia/Ho_Chi_Minh",
  "cron_expr": "0 9 * * *",
  "max_runs_per_day": 1,
  "status": "active",
  "results": null,
  "campaign": "morning-promo-q3",
  "created_by": "alice",
  "status_changed_by": "alice",
  "created_at": "2026-08-08T12:00:00Z",
  "updated_at": "2026-08-08T12:00:00Z"
}
```

`400` on validation failure (no `text`/`media`, no `channels` unless `save_as_draft`, an unregistered channel, or an invalid `cron_expr`).

#### Update a post

```
PATCH /v1/posts/:id
Content-Type: application/json
```
Only while `status` is `draft` or `pending_review` (`400` otherwise - cancel a scheduled/active post first if it needs changes). Every field is optional; omitted fields are left unchanged:
```json
{
  "title": "August flash sale v2",
  "channels": ["telegram", "twitter"],
  "scheduled_at": "2026-08-10T09:00:00Z"
}
```
Editing a `pending_review` post demotes it back to `draft` — the reviewed content no longer matches what was approved, so it needs resubmitting. `media`/`channels` replace the existing list wholesale when provided (no partial add/remove).

#### Submit for review

```
POST /v1/posts/:id/submit
```
`draft → pending_review`. Requires `channels` and a schedule (`scheduled_at` or `cron_expr`) to already be set — use `PATCH` first if the draft doesn't have them yet. `400` otherwise.

#### Approve / reject

```
POST /v1/posts/:id/approve
POST /v1/posts/:id/reject
Content-Type: application/json

{"changed_by": "bob", "reason": "typo in the headline"}   -- reject; approve only takes changed_by
```
Both require `status = pending_review`. Approve moves to `scheduled` (one-off) or `active` (recurring) and records `changed_by` as `approved_by`. Reject moves back to `draft` and records `reason` in `reject_reason` for the author to see. See "Actor tracking" below for `changed_by`.

#### List posts

```
GET /v1/posts?limit=20&offset=0&status=pending_review&campaign=morning-promo-q3
```
`status`/`campaign` are both optional filters — omit either to not filter on it. Handy for a review queue (`?status=pending_review`) or a campaign report (`?campaign=morning-promo-q3`).
```json
{"posts": [ { "...": "model.Post" } ]}
```

#### Get one post

```
GET /v1/posts/:id
```
Returns the `model.Post`, or `404` if it doesn't exist.

#### Publish now

```
POST /v1/posts/:id/publish
Content-Type: application/json

{"changed_by": "alice"}
```
Force an immediate run regardless of `scheduled_at`/`cron_expr` — works on any post, not just `scheduled`/`active` ones (unlike the automated scheduler path, which never touches anything else). For a recurring post this **bypasses** `max_runs_per_day`/`max_runs_per_month` (an explicit manual trigger always runs) but still counts toward them. Returns the updated `model.Post`.

#### Cancel

```
POST /v1/posts/:id/cancel
Content-Type: application/json

{"changed_by": "alice"}
```
Valid from `draft`, `pending_review`, `scheduled`, or `active`; `204` on success, `400` if already terminal (`published`/`failed`/`cancelled`).

### Actor tracking

`created_by` (set at creation), `approved_by` (set by `approve`), and `status_changed_by` (set by *every* status-changing action — submit/approve/reject/cancel/publish, plus the scheduler itself, which records `"scheduler"`) are free-form strings the caller supplies via `changed_by`/`created_by` in the request body. This service has no authentication of its own, so it trusts whatever identity its caller passes in — if that matters to you, enforce identity upstream (a gateway, a reverse proxy) and forward the verified username through.

### `model.Post` status values

| Status | Applies to | Meaning |
|---|---|---|
| `draft` | either | content saved, not yet submitted for review (or bounced back by a reject) |
| `pending_review` | either | submitted, waiting on `approve`/`reject` |
| `scheduled` | one-off | approved/direct-created, waiting for `scheduled_at` |
| `publishing` | either | a short-lived lock held only while a publish attempt is actually in flight - see "Preventing duplicate posts" |
| `published` | one-off | terminal: every channel succeeded |
| `failed` | one-off | terminal: at least one channel failed |
| `active` | recurring | running on `cron_expr`; per-run outcome is in `results`, not the status |
| `cancelled` | both | stopped by `POST /:id/cancel`, terminal |

### Preventing duplicate posts

Three separate mechanisms guard against sending the same content twice:

1. **Status gate on the automated path.** `PublishDuePosts` only ever queries for `status = scheduled` (one-off) or `status = active` (recurring), and re-checks that status in Go before dispatching. A draft, a pending-review post, or anything else can never be auto-sent - only `PublishNow` can force a post through regardless of status, and that's an explicit manual action.
2. **A `publishing` lock around every dispatch.** Immediately before calling out to any channel, the post is flipped to `publishing` and persisted; only after the attempt finishes does it move to its real outcome (`published`/`failed`, or back to `active` for a recurring post). Since the automated queries only ever return `scheduled`/`active` posts, one already sitting in `publishing` is invisible to a second, overlapping run. The scheduler additionally wraps its cron job in `robfig/cron`'s `SkipIfStillRunning`, so a slow tick (e.g. a flaky social API) can't have the *next* tick start a second, concurrent `PublishDuePosts` in the same process.
3. **Per-channel skip on retry.** If a post partially fails (e.g. Telegram succeeds, Twitter times out) and is later retried — via `PublishNow`, or a future scheduler pass for a post that somehow re-entered `scheduled` — channels that already have a successful `results` entry are **not** re-published; their existing `external_id`/`published_url` is carried forward untouched. Recurring posts are the one exception: every firing is an intentional new publish, not a retry, so this skip never applies to them.

None of this covers **multiple replicas** of this service — see "Known limitations".

`results` always reflects the **most recent** publish attempt only (not a history log) — `{channel, external_id, published_url, error, published_at}` per channel targeted. `published_url` is best-effort: derived from `external_id` for Twitter/Facebook (a canonical status/post URL), left blank for Telegram (no permalink is derivable without a public channel username, which isn't part of config). `reject_reason` similarly only holds the *most recent* rejection's feedback, cleared on the next submit/approve.

### Facebook webhook (only when `FACEBOOK_ENABLED=true`)

```
GET  /webhook/facebook   # Messenger Platform verification handshake (hub.challenge)
POST /webhook/facebook   # inbound messages, X-Hub-Signature-256 verified against FACEBOOK_APP_SECRET
```
Not meant to be called directly — point Meta's App Dashboard webhook config at this URL.

## Development

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

## Known limitations

- Twitter/Facebook publishing supports at most one image per post; no chunked/video upload, no multi-photo galleries.
- No cross-process locking: the `publishing` status lock and `SkipIfStillRunning` only protect a single running instance. Running multiple replicas of this service still risks the scheduler double-publishing the same due post - a genuine fix needs `SELECT ... FOR UPDATE SKIP LOCKED` or leader election, neither of which is implemented.
- `migrations/*.sql` are applied by hand — there's no migration runner or `AutoMigrate` wired up, so schema drift between environments is on you to manage.
- No auth: `created_by`/`approved_by`/`status_changed_by` are trusted as plain strings from whoever calls the API.
