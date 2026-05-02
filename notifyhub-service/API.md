# NotifyHub API Reference

**Base URL:** `http://localhost:8080`  
**Auth:** `X-API-Key: <secret>` or `Authorization: Bearer <secret>`  
**Content-Type:** `application/json`

All responses include `ts` (Unix timestamp) and `request_id` fields.

---

## Public Endpoints

### GET /health
```json
{ "status": "ok", "ts": 1717200000, "version": "1.0.0" }
```

### GET /metrics
Prometheus text format metrics.

---

## Channels

A Channel configures one notification provider. The `type` field determines which underlying driver is used. Channel `config` is reserved for future per-channel overrides.

### POST /api/v1/channels
```json
{
  "name": "primary-email",
  "type": "email",
  "is_active": true,
  "config": {}
}
```
`type` values: `email` | `sms` | `firebase`

**Response 201:**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "primary-email",
    "type": "email",
    "is_active": true,
    "created_at": "2025-06-01T08:00:00Z",
    "updated_at": "2025-06-01T08:00:00Z"
  },
  "ts": 1717200000,
  "request_id": "abc-123"
}
```

### GET /api/v1/channels
Query params: `type=email|sms|firebase`, `active=true|false`

### GET /api/v1/channels/:id
### PUT /api/v1/channels/:id
### DELETE /api/v1/channels/:id

---

## Data Sources

A DataSource defines an external REST API to call before rendering each notification. The fetched JSON is merged into the template context.

### POST /api/v1/data-sources
```json
{
  "name": "user-api",
  "url": "https://api.example.com/users/current",
  "method": "GET",
  "headers": {
    "Accept": "application/json",
    "X-App-Version": "2"
  },
  "auth_type": "bearer",
  "auth_config": {
    "token": "eyJhbGciOiJIUzI1NiJ9..."
  },
  "json_path": "$.data",
  "is_active": true
}
```

**auth_type values:**

| Type | Required auth_config fields |
|------|----------------------------|
| `none` | — |
| `bearer` | `token` |
| `basic` | `username`, `password` |
| `apikey` | `key`, `header` (default: `X-API-Key`), or `param` for query string |

**json_path:** Dot-notation path into the response JSON.  
- `""` → use entire response  
- `$.data` → extract `.data` field  
- `$.data.users` → extract `.data.users`

### GET /api/v1/data-sources
### GET /api/v1/data-sources/:id
### PUT /api/v1/data-sources/:id
### DELETE /api/v1/data-sources/:id  (soft delete via is_active=false)

---

## Templates

Templates use Go's `text/html/template` syntax. Variables come from two sources merged together:
1. `static_payload` on the job
2. Data fetched from the job's DataSource

### POST /api/v1/templates
```json
{
  "name": "welcome-email",
  "channel": "email",
  "subject": "Welcome to {{.Company}}, {{.UserName}}!",
  "body": "<h1>Hi {{.UserName}}</h1><p>Your account at <strong>{{.Company}}</strong> is ready.</p>{{if .IsPremium}}<p>You have <strong>Premium</strong> access.</p>{{end}}",
  "variables": ["UserName", "Company", "IsPremium"],
  "is_active": true
}
```

**Template syntax reference:**

```
{{.FieldName}}                        Simple variable substitution
{{if .IsActive}}...{{end}}            Conditional
{{if .IsActive}}...{{else}}...{{end}} If-else
{{range .Items}}{{.}}{{end}}          Iterate over a slice
{{range .Items}}{{.Name}}{{end}}      Iterate over structs
{{.Field | printf "%.2f"}}            Pipe to built-in function
```

> **Note:** When using a DataSource, fetched JSON fields are merged at the top level.  
> E.g., if the API returns `{"name":"Alice","score":99}`, use `{{.name}}` and `{{.score}}`.

### GET /api/v1/templates?channel=email
### GET /api/v1/templates/:id
### PUT /api/v1/templates/:id  (also recompiles the template cache)
### DELETE /api/v1/templates/:id  (soft delete)

---

## Jobs

A Job ties Channel + Template + optional DataSource to a schedule and a list of recipients.

### POST /api/v1/jobs — Cron
```json
{
  "name": "Daily Sales Report",
  "description": "Emails daily sales summary every weekday at 8AM UTC",
  "channel_id": "CHANNEL_UUID",
  "template_id": "TEMPLATE_UUID",
  "data_source_id": "DATASOURCE_UUID",
  "schedule_type": "cron",
  "cron_expr": "0 8 * * 1-5",
  "recipients": [
    "sales@company.com",
    "manager@company.com"
  ],
  "static_payload": {
    "Company": "Acme Corp",
    "ReportDate": "today"
  },
  "max_runs": 0
}
```

### POST /api/v1/jobs — One-time
```json
{
  "name": "Launch Day Announcement",
  "channel_id": "CHANNEL_UUID",
  "template_id": "TEMPLATE_UUID",
  "schedule_type": "once",
  "run_at": "2025-12-01T10:00:00Z",
  "recipients": ["+84901234567", "+84909876543"],
  "static_payload": {
    "ProductName": "SuperApp 2.0"
  }
}
```

### POST /api/v1/jobs — Interval
```json
{
  "name": "Heartbeat Push",
  "channel_id": "FIREBASE_CHANNEL_UUID",
  "template_id": "PUSH_TEMPLATE_UUID",
  "schedule_type": "interval",
  "interval_sec": 1800,
  "recipients": ["FCM_DEVICE_TOKEN_HERE"],
  "static_payload": {
    "title": "System Update",
    "body": "Your app is up to date."
  },
  "max_runs": 48
}
```

**Fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `name` | ✅ | Human-readable name |
| `channel_id` | ✅ | Channel UUID |
| `template_id` | ✅ | Template UUID |
| `data_source_id` | ❌ | DataSource UUID (fetched before each run) |
| `schedule_type` | ✅ | `cron` \| `once` \| `interval` |
| `cron_expr` | cron only | Standard 5-field cron expression |
| `run_at` | once only | RFC3339 datetime, must be in future |
| `interval_sec` | interval only | Positive integer seconds |
| `recipients` | ✅ | Array of email / phone / FCM token |
| `static_payload` | ❌ | Key-value template variables |
| `max_runs` | ❌ | Stop after N runs (0 = unlimited) |

### GET /api/v1/jobs
Query params: `status=active|paused|completed|failed`, `page=1`, `page_size=20`

**Response:**
```json
{
  "data": [...],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "ts": 1717200000
}
```

### GET /api/v1/jobs/:id
### PUT /api/v1/jobs/:id   (re-registers in scheduler automatically)
### DELETE /api/v1/jobs/:id

### POST /api/v1/jobs/:id/pause
Removes the job from the scheduler. History and config are preserved.
```json
{ "data": { "status": "paused" }, "ts": 1717200000 }
```

### POST /api/v1/jobs/:id/resume
Re-registers a paused job in the scheduler.
```json
{ "data": { "status": "active" }, "ts": 1717200000 }
```

### POST /api/v1/jobs/:id/trigger
Fires the job immediately, bypassing its schedule. Useful for testing.
```json
{ "data": { "triggered": true, "job_id": "UUID" }, "ts": 1717200000 }
```

---

## Notification History

Read-only log of every dispatch attempt.

### GET /api/v1/history
Query params: `job_id`, `status=pending|sent|failed|retrying`, `page`, `page_size`

**Response:**
```json
{
  "data": [
    {
      "id": "UUID",
      "job_id": "UUID",
      "channel_type": "email",
      "recipient": "admin@example.com",
      "subject": "Daily Report - Acme Corp",
      "body": "<h1>Hi Admin</h1>...",
      "status": "sent",
      "retry_count": 0,
      "error": "",
      "sent_at": "2025-06-01T08:00:03Z",
      "created_at": "2025-06-01T08:00:01Z"
    }
  ],
  "total": 1240,
  "page": 1,
  "page_size": 20
}
```

---

## Scheduler Status

### GET /api/v1/scheduler/status
```json
{
  "scheduled_jobs": ["job-uuid-1", "job-uuid-2"],
  "count": 2,
  "ts": 1717200000
}
```

---

## Cron Expression Reference

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-7, 0=Sunday)
│ │ │ │ │
* * * * *
```

| Expression | Meaning |
|------------|---------|
| `* * * * *` | Every minute |
| `0 * * * *` | Top of every hour |
| `0 8 * * *` | Every day at 08:00 UTC |
| `0 8 * * 1-5` | Weekdays at 08:00 UTC |
| `0 8 * * 1` | Every Monday at 08:00 UTC |
| `0 0 1 * *` | First day of each month |
| `*/15 * * * *` | Every 15 minutes |
| `0 9,18 * * 1-5` | 9AM and 6PM on weekdays |
| `0 0 * * 0` | Every Sunday at midnight |

---

## Error Responses

```json
{
  "error": "job name is required",
  "ts": 1717200000,
  "request_id": "abc-123"
}
```

| HTTP Status | Meaning |
|-------------|---------|
| `400` | Validation error or bad request body |
| `401` | Missing or invalid API key |
| `404` | Resource not found |
| `429` | Rate limit exceeded (default: 100 req/s per IP) |
| `500` | Internal server error |
