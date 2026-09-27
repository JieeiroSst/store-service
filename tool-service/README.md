# tool-service

AI QC assistant: reads a task document (PDF / Word / text), decides what to test, writes API test cases, runs them,
load-tests the endpoints and returns a pass/fail gate. Backed by Ollama. Web UI at `/`.

## How correctness is protected

A language model cannot guarantee that a test it wrote is right, and for money that is not acceptable.
AI output is therefore never allowed to block (or pass) a release on its own:

| Layer | What it does |
|---|---|
| **Exact arithmetic** | Expected values are written as `"calc:19.99*3"` and computed by code with exact decimals; responses are compared as exact numbers (no float rounding). |
| **Grounding** | Every case must quote the task document verbatim (`spec_quote`); a case whose quote is not in the document is untrusted. |
| **Consistency** | The suite is generated several times independently (`consistency_runs`, default 3); only assertions a majority agree on can be trusted. |
| **Verification** | An independent pass confirms each expectation is implied by its quote. Any doubt means untrusted. |
| **Human approval for money** | Money-related cases (amount, price, balance, payment, tax, ...) become blocking only after a person approves them into a **golden suite**. |
| **Golden suite** | Approved cases run with **no AI at all**: deterministic, the gate to trust in CI. |

Gate status: `passed`, `failed` (a trusted case or a performance target failed) or `needs_review` (nothing trusted failed,
but some cases could not be trusted, or money cases await approval). An untrusted case failing is a review item, never proof of a bug.

Measure it, don't assume it: `GET /api/v1/suites/<name>` reports how often reviewers accepted the model's proposals with a
95 % Wilson lower bound (`meets_99_percent`). A "99 % correct" claim is only supported once that bound reaches 0.99,
which takes hundreds of reviewed cases.

## What is tested

| Area | How |
|---|---|
| API functional / regression | AI-proposed cases (trusted only when grounded, consistent, verified) + human-approved golden suite |
| Requirement coverage | Normative statements of the document (must / shall / phải / tối đa ...) not quoted by any test are listed, and a second generation round covers them |
| Performance | Latency percentiles and throughput vs targets; **stress** ramp to the breaking point; **soak** runs up to 1 h that detect latency/error drift (background jobs) |
| Security, passive | TLS, headers, cookies, CORS, exposed files (`.env`, `.git`, pprof...), anonymous access, error leakage |
| Security, active + exploratory | Injection probes (XSS, SQL error, SSTI, path traversal), malformed-input fuzzing, negative/overflow money values, authorization matrix between identities, edge-case values through the site's own forms |
| Web pages | Crawl: broken links, duplicate ids, mixed content, alt text, form labels, lang/title/viewport, headings |
| Browser (real Chromium) | Desktop/tablet/mobile layout and overflow, WCAG contrast, tap-target size, keyboard operation (focus order, traps, focus indicator, click-only elements), accessibility tree (names, landmarks, headings: what a screen reader consumes), localization (Accept-Language, RTL, untranslated keys), console errors, failed requests, **visual regression** against a stored baseline |
| Database | PostgreSQL / MySQL, read-only transactions: missing keys, orphan rows, money columns (float type, negatives) and your own SQL business rules |
| Async and queues | Trigger, then poll until the effect is visible (time to consistency) and replay with the same Idempotency-Key; RabbitMQ backlog / no consumers / dead-letter queues via the management API |
| Mobile: static (APK, IPA) | No device needed. Android: debuggable / testOnly, backup, cleartext, target SDK, exported components without permission, risky permissions, signing, 64-bit libraries. iOS: ATS, usage descriptions checked against the APIs the binary really uses (the classic App Store rejection), privacy manifest, provisioning (debug entitlement, expiry). Both: embedded secrets and keys |
| Mobile: Android device or emulator | Over adb: install, cold start time, memory, jank, crashes/ANRs/native crashes from logcat, monkey (random input) test, missing accessibility labels and touch targets under 48 dp, then rotation, 200% font, dark mode, RTL locale and backgrounding, with screenshots; device settings are restored afterwards |
| Mobile: iOS simulator | Over `xcrun simctl` (macOS host with full Xcode only): install a simulator `.app`, launch, screenshots, crash reports, dark mode, largest Dynamic Type, RTL locale, deep links, fault-level logs |
| UI end-to-end, cross-browser | Playwright script and config for Chromium, Firefox, WebKit and mobile devices are **generated** (`e2e=true`); Chromium is the only browser executed inside this service |

Every analysis response carries a `scope` list stating this.

### Endpoints
| Endpoint | Purpose |
|---|---|
| `POST /api/v1/qa/scan` | passive security scan + web crawl (no AI) |
| `POST /api/v1/qa/browser` | browser audit `{base_url, urls, viewports, locales, name, update_baseline}` (use `?async=true`) |
| `POST /api/v1/qa/database` | `{connection, assertions:[{name, sql, expect, value}]}` |
| `POST /api/v1/qa/async` | `{base_url, flow:{trigger, poll, timeout_sec, idempotency}}` |
| `POST /api/v1/qa/queue` | `{broker, max_ready}` |
| `POST /api/v1/qa/active` | `{base_url, active:true, endpoints, identities, resources, include_writes, explore_forms}` |
| `POST /api/v1/jobs/stress`, `/jobs/soak` | long load tests as background jobs |
| `GET /api/v1/jobs`, `/jobs/:id`, `POST /jobs/:id/cancel` | job status and results (kept in memory) |
| `POST /api/v1/qa/mobile/static` | multipart `file` = `.apk` or `.ipa` (max 200 MB) |
| `POST /api/v1/qa/mobile/android` | multipart: optional `file` (APK), `package`, `serial`, `monkey_events`, `seed`, `scenarios`, `uninstall` (use `?async=true`) |
| `POST /api/v1/qa/mobile/ios` | multipart: optional `file` (zipped simulator `.app`), `bundle_id`, `device`, `deep_links`, `scenarios` |
| `GET /api/v1/qa/mobile/devices` | Android devices seen by adb and available iOS simulators |
| `GET /api/v1/artifacts/:file` | screenshots and diff images |

All accept `?fail_http=true` (HTTP 422 when the gate fails) and return `gate.status` = `passed` / `failed` / `needs_review`.

### Safety rules
* Database, queue and active-security endpoints are **refused unless `API_TOKEN` is configured**.
* Database and RabbitMQ credentials live on the server (`DB_CONNECTIONS`, `RABBITMQ_CONNECTIONS`), never in API requests.
  Assertions must be a single `SELECT`/`WITH` and run in a read-only transaction with a statement timeout.
* Active testing requires `"active": true` in the request, a non-empty `TARGET_ALLOWLIST`, and `include_writes` for any request that can change data. It sends detection payloads only and stops at a request budget (default 300).
* Redirects and browser sub-requests that leave `TARGET_ALLOWLIST` are refused.
* Only test systems you own or are authorised to test.

### Mobile testing setup
* **Android**: the image contains `adb`. Point it at a device with `ANDROID_ADB_SERVER_ADDRESS` / `ANDROID_ADB_SERVER_PORT` (an emulator container, a device farm, or a host running `adb -a nodaemon server`). Locally, connect a device or start an emulator. Set `ADB_PATH` if adb is not on `PATH`.
* **iOS**: only from a macOS host with full Xcode (`xcode-select -s /Applications/Xcode.app`), by running this service there. The Linux image answers 501. Upload a `.app` built for the simulator, zipped; a device `.ipa` cannot run in the simulator (use the static analysis for it).
* Package names, bundle ids, serials, device names and deep links are validated against strict patterns; only fixed adb/simctl commands are issued, never user-supplied shell text.

### Browser in containers
The image is based on `chromedp/headless-shell` (Chromium built for headless use; Alpine's chromium crash-looped in some VMs) plus fonts. Set `CHROME_PATH` if the browser lives elsewhere, `CHROME_NO_SANDBOX=true` when running as root,
`CHROME_FLAGS` for extra flags, `CHROME_DEBUG=true` to see the browser's output. Kubernetes needs a memory-backed `/dev/shm` (the chart adds one).

## Workflow
1. Analyze: `POST /api/v1/tasks/analyze` (multipart: `file`, `base_url`, `suite`) or use the web UI.
2. A person reviews and approves the correct cases:
   `POST /api/v1/suites/<name>/approve` `{"experience_id":"...","approve":["case"],"reject":["other"]}`
   (or hand-written: `{"cases":[...]}`).
3. In CI run the golden suite (no AI): `POST /api/v1/suites/<name>/run` `{"base_url":"..."}`;
   `?format=junit` for test reports, `?fail_http=true` to get HTTP 422 on failure so `curl --fail` breaks the build.

```yaml
# GitHub Actions step (after deploying to staging)
- name: Golden suite gate
  run: |
    curl --fail-with-body -sS -X POST "$QC_URL/api/v1/suites/payments/run?fail_http=true" \
      -H "Authorization: Bearer $QC_API_TOKEN" -H 'Content-Type: application/json' \
      -d "{\"base_url\":\"$STAGING_URL\"}"
```

## `POST /api/v1/tasks/analyze` fields
| Field | Meaning |
|---|---|
| `file` / `spec` | task document (`.pdf`, `.docx`, `.txt`, `.md`) or plain text |
| `base_url` | deployed API under test; without it only plan and cases are produced |
| `suite` | golden suite: approved cases always run and stay trusted |
| `consistency_runs` | independent generations to compare, 1 to 5 (default 3) |
| `min_pass_rate` | required pass rate among trusted cases (default `1.0`) |
| `unreviewed` | `block` (default) or `warn`: whether `needs_review` fails the gate |
| `security`, `crawl`, `max_pages`, `e2e`, `fill_gaps` | product checks (`security`/`crawl` default `true`), crawl size (default 20), generate Playwright script, cover gaps (default `true`) |
| `max_error_rate`, `benchmark` | load-test error budget (default 1 %), `false` to skip |
| `code`, `language` | optional source to review against the task |
| `format`, `fail_http` | `json` or `junit`; `true` returns HTTP 422 when the gate fails |

## Configuration (environment)
`OLLAMA_URL`, `OLLAMA_MODEL`, `OLLAMA_EMBED_MODEL`, `API_TOKEN` (Bearer / `X-API-Key`), `TARGET_ALLOWLIST`,
`DB_CONNECTIONS`, `RABBITMQ_CONNECTIONS`, `CHROME_PATH`, `CHROME_FLAGS`, `LEARNING_DATA_DIR`, `CONSISTENCY_RUNS`, `AUTO_LEARN_INTERVAL`, `EVOLVE_EVERY_APPROVED`, `HTTP_PORT`.
Deploy: `chart/tool-service`, workflows `.github/workflows/tool-service-cicd.yml` and `go.yml`.
