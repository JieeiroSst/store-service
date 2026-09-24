# customer-relationship-service

CRM API (hexagonal architecture, `go.uber.org/fx` for dependency injection).
Configuration is via environment variables: see `.env.example`.

## Endpoints (`/api/v1`, plus `GET /health`)

CRUD (`POST /x`, `GET /x`, `GET /x/:id`, `PUT /x/:id`, `DELETE /x/:id`) for
`campaigns`, `leads`, `accounts`, `contacts`, `campaign-members`, `cases`,
`contracts`, `account-contact-roles`, `opportunities`,
`opportunity-contact-roles`.

Lists take `offset`, `limit` (max 200), `q` (text search) and exact-match
filters, and return the total in the `X-Total-Count` header:

| resource | filters |
| --- | --- |
| leads | `status`, `source` |
| contacts | `account_id` |
| campaign-members | `campaign_id`, `lead_id`, `contact_id` |
| cases | `contact_id`, `status`, `priority` |
| contracts | `account_id`, `contract_status` |
| opportunities | `account_id`, `opportunity_stage` |

Workflows:

- `POST /leads/:id/convert` - creates an account + contact (+ opportunity with `create_opportunity`) in one transaction; `409` if already converted.
- `POST /opportunities/:id/stage` `{"stage": "..."}` - prospecting, qualification, proposal, negotiation, won, lost. Won/lost are final.
- `POST /contracts/:id/{submit,approve,reject}` - draft -> pending -> approved/rejected. A contract already past its `end_date` cannot be submitted or approved until it gets a new one.
- `GET /contracts/:id/signing-payload?signed_by=&end_date=[&signing_time=]` - the exact bytes to sign for a detached signature.
- `GET /contracts/:id/document?signed_by=&end_date=&signing_time=` - the renewal PDF to sign for a PAdES signature.
- `POST /contracts/:id/sign` - the only way to reopen an `expired` contract; see "Digital signatures". `422` for a bad signature or certificate, `502` when the configured time-stamp authority fails.
- `GET /contracts/:id/signed-document` - the signed PDF kept by the last PAdES renewal (the newest file of kind `renewal`, see "Contract files").
- `POST /cases/:id/close`.
- `GET /accounts/:id/overview` - account with its contacts, contracts and opportunities.
- `GET /reports/summary` - counts, leads by status, open cases, pipeline by stage.

### Contract lifecycle (Temporal)

With `TEMPORAL_ADDRESS` set (the chart points it at the `temporal` chart's `temporal-service:7233`), every contract that can expire has a durable workflow, `ContractLifecycle` (workflow id `contract-<id>`, task queue `crm-contract-lifecycle`), run by a worker inside this service:

- while the contract is draft, pending or approved and has an `end_date`, the workflow sleeps until that date and then closes the contract (`expired`);
- while it is **approved** it also sends a reminder at each offset in `TEMPORAL_REMINDER_DAYS` (default 30, 7 and 1 days) before the end date; reminders already in the past when it starts looking at an end date are skipped, not sent late;
- once expired it waits, with no timer, until the contract is renewed with `sign`;
- it finishes when the contract is rejected or deleted, and continues-as-new every 200 wake-ups to keep its history small.

The database stays the source of truth and the HTTP transitions (`submit`, `approve`, `reject`, `sign`, `PUT`, `DELETE`) stay synchronous, so they keep returning `409`/`422` directly. After each change the service sends a `state-changed` signal (SignalWithStart, so it also starts the workflow if needed) and the workflow re-reads the contract. Expiry passes the workflow's clock to the database, so a skewed application clock cannot make the two disagree, and an expire that finds nothing to do backs off 30 seconds instead of spinning.

At startup the worker resyncs every draft, pending and approved contract that has an end date, so contracts created before Temporal was switched on get a workflow. Calls to Temporal are best-effort with a 3 second timeout: if Temporal is down, contract requests still succeed and the cron reconciler below and the next startup resync catch up. Lead conversion, opportunity stages and case closing are single-transaction operations and are not workflows.

Tested against the workflow test environment (time-skipping) and against a live Temporal server 1.16.2, the version the `temporal` chart deploys. Run the live test with `TEMPORAL_TEST_ADDRESS=localhost:7233 go test ./internal/adapter/primary/temporalworker`.

Contract expiry (reconciler): a cron job (`CONTRACT_EXPIRY_CRON`, default hourly, also run once at startup) sets draft/pending/approved contracts whose `end_date` has passed to `expired`. It is a single idempotent UPDATE, so running it from several replicas is safe. Expired contracts stay closed: workflow steps and `PUT` cannot reopen them, only `sign`.

## Authentication and roles

Every `/api/v1` request is authenticated by one of:

- **An OIDC bearer token** (`Authorization: Bearer ...`) from `OIDC_ISSUER` (the `keycloak` chart's realm): RS/PS/ES-signed JWTs only, checked against the provider's JWKS (`OIDC_JWKS_URL`, default `<issuer>/protocol/openid-connect/certs`; keys are cached and refreshed, a token signed with an unknown key triggers at most one refresh a minute). `iss`, `exp` and `sub` are required, `aud` is checked when `OIDC_AUDIENCE` is set, `none`/HMAC tokens are refused. The caller is the token's `sub` (display name from `preferred_username`, `name` or `email`). While the provider cannot be reached the answer is `503`, not `401`.
- **The API key** (`X-API-Key`, constant-time comparison): a service caller (`api-key`) whose role is `API_KEY_ROLE` (default `admin`).
- With neither `API_KEY` nor `OIDC_ISSUER` set the API is **open**, every caller is an administrator, and a warning is logged at startup.

Roles are read from the token at `OIDC_ROLES_CLAIM` (default `realm_access.roles`, Keycloak's realm roles) and only those starting with `ROLE_PREFIX` (default `crm-`) count: `crm-viewer`, `crm-staff`, `crm-manager`, `crm-admin`. The highest one applies. A person with none of them is authenticated but may do nothing with contract files (`403`).

| | viewer | staff | manager | admin |
| --- | --- | --- | --- | --- |
| list, get, download, history of files | yes | yes | yes | yes |
| upload (and replace their own files) | | yes | yes | yes |
| delete | | their own uploads | any | any |
| replace someone else's file | | | yes | yes |
| read the contract-wide audit trail | | | yes | yes |

Files of kind `legal` are restricted: below manager they are invisible (not listed, and `404` rather than `403` so their existence is not revealed) and cannot be uploaded. Files of kind `renewal` can never be uploaded or deleted through the API, by anyone. The authorization lives in the use case (`internal/domain/auth`, checked in `internal/application/contract_files.go`), not in the HTTP layer. It applies to contract files; other endpoints only require a valid credential.

## Contract files

Every contract can carry any number of files, each of one **kind**:

| kind | what it is | formats |
| --- | --- | --- |
| `contract` | the contract itself (hợp đồng) | pdf, docx, doc |
| `appendix` | phụ lục | pdf, docx, doc |
| `renewal` | the signed renewal appendix (kept by the service) | pdf, **system only** |
| `scan` | scan of the signed paper contract | pdf, jpg, png |
| `legal` | legal papers: business registration, power of attorney, ... (**manager only**) | pdf, jpg, png |
| `acceptance` | biên bản nghiệm thu, thanh lý | pdf, docx, doc, jpg, png |
| `payment` | payment documents, invoices | pdf, xlsx, xls, jpg, png |
| `attachment` | other attachments | pdf, docx, doc, xlsx, xls, jpg, png |
| `other` | anything else | same as `attachment` |

`GET /contract-file-kinds` returns this list.

- `POST /contracts/:id/files` - multipart form: `file` (required), `kind` (required), `description`, `replaces` (id of the file this is a new version of). The uploader is the authenticated caller. `201` with the file's metadata.
- `GET /contracts/:id/files[?kind=][&all_versions=true][&offset=&limit=]` - list, oldest first, only the current version of each file unless `all_versions=true`; total in `X-Total-Count`.
- `GET /contracts/:id/files/:file_id` - metadata (name, kind, content type, format, size, `sha256`, description, uploader, source `upload` or `system`, version, `latest`, scan result, timestamps). The storage key is never exposed.
- `GET /contracts/:id/files/:file_id/download` - the bytes, always as an attachment with the content type we recognised, `X-Content-Type-Options: nosniff` and `Cache-Control: private, no-store`.
- `GET /contracts/:id/files/:file_id/history` - every version of the file (deleted ones included) and every recorded action on any of them.
- `GET /contracts/:id/file-events[?file_id=][&action=]` - the audit trail of all the contract's files (managers).
- `DELETE /contracts/:id/files/:file_id` - hides the file (`204`); the stored bytes are kept.

### What an upload must pass

In this order, and a refusal is on the audit trail (action `rejected`):

1. **Authorization** (`403`) and an existing contract (`404`).
2. **Type**, by content rather than what the client says: the extension must be an accepted one, the bytes must be that format (PDF header; PNG/JPEG headers that parse; DOCX/XLSX zips holding the Office manifest and `word/document.xml` or `xl/workbook.xml`; legacy OLE for doc/xls), and the kind must accept the format: `415`. Unknown kind `400`; over `CONTRACT_FILE_MAX_MB` (default 25) `413`.
3. **Active content** (`422`): PDFs are parsed and refused if they contain JavaScript, launch or submit-form actions, embedded files, rich media, XFA forms and the like (a second, raw-byte pass catches what the parser cannot reach); DOCX/XLSX are refused if they carry macros (`vbaProject.bin`) or ActiveX controls, or expand past 512 MB. `PDF_POLICY`: `strict` (default) also refuses PDFs that are encrypted or cannot be parsed, since they cannot be checked; `lenient` refuses only what it finds; `off` skips the PDF checks. Legacy `.doc`/`.xls` cannot be inspected here; they rely on the virus scan.
4. **Virus scan** with ClamAV (`CLAMAV_ADDRESS`, the `clamav` chart's `clamav-svc:3310`): the file is streamed to clamd; a detection is refused (`422`, naming the signature). If clamd cannot be reached: with `CLAMAV_REQUIRED=true` (default) the upload is refused (`502`); otherwise it is accepted with `scan_status: skipped`. Without an address every file has `scan_status: skipped`. Files the service itself produces are not scanned (`scan_status: system`). Raise clamd's `StreamMaxLength` (the chart sets 100M) above `CONTRACT_FILE_MAX_MB`. ClamAV's EICAR signature only matches a file that is the test string itself.

File names are reduced to a display name (no directories or control characters, at most 200 characters, Unicode kept); object keys are `contracts/<contract id>/<kind>/<date>-<random>.<ext>` and never contain the client's name. The size limit does not apply to files the service itself produces.

### Versions and history

Upload with `replaces=<file id>` to add a new version: it must be the same contract and kind, the current version, and the caller must be its uploader or a manager. The new file joins the same lineage with the next version number and becomes the current one; older versions stay downloadable (`all_versions=true`). Deleting the current version makes the previous one current again. Signed renewals chain the same way automatically: each renewal of a contract is the next version of the previous one.

Every action is recorded in `crm_contract_file_event`, which is only ever appended to: `uploaded`, `new_version`, `downloaded`, `deleted`, `rejected` (an upload the checks refused, with the reason and no file id) and `denied` (an action the caller's role does not allow). Each entry has the time, the actor (id, display name, roles, whether it is a service key, source IP), the file (id, lineage, version, name, kind) and a detail. Uploads and deletes are written in the same transaction as the change, so a file never exists without its entry (and the uploaded bytes are removed if the transaction fails). Downloads, refusals and denials are best-effort: if the entry cannot be written, a warning is logged and the request still completes.

Storage is as for the signed PDFs: with `MINIO_ENDPOINT` set (the chart points it at the `minio` chart's `minio-svc:9000`) the bytes go to the `MINIO_BUCKET` bucket and the `crm_contract_file` row keeps the key, the size and the SHA-256; without it they stay in the row. Every download re-hashes what it read and returns `500` if it no longer matches, so a file edited in the bucket is detected rather than served. A row saved with a MinIO key cannot be served if MinIO is later switched off.

Not covered: content disarm (files are refused, never cleaned), inspection of legacy `.doc`/`.xls` beyond the virus scan, per-contract permissions (roles are global: a staff member sees every contract's files), purging deleted files' bytes from the bucket, and retention limits on the audit trail. The earlier `crm_contract_document` table is no longer used: run a one-off copy into `crm_contract_file` (kind `renewal`) if you already have rows in it.

## Contract document (PDF)

`GET /contracts/:id/document` renders a "Phụ lục gia hạn hợp đồng" in the layout of a Vietnamese contract: quốc hiệu and tiêu ngữ, số văn bản, căn cứ pháp lý, "Bên A" (from the `COMPANY_*` settings) and "Bên B" (the contract's account: name, `billing_address`, `tax_code`, `phone`, `representative_title`; the representative is the signer), the renewal terms (Điều 1 gia hạn, Điều 2 các nội dung khác, Điều 3 hình thức và hiệu lực, Điều 4 giải quyết tranh chấp), two signature blocks and a footer with page numbers and a short verification code. The contract's `contract_number` and `contract_title` fill the số hợp đồng and subject; dates are shown in Vietnam time (UTC+7). Set in Tinos (Apache 2.0, metric-compatible with Times New Roman, full Vietnamese coverage), embedded in the binary. The output is byte-for-byte reproducible, which the PAdES check below relies on: change the contract or its account between `document` and `sign` and the signed file no longer matches (`422`).

The legal wording is a template: it cites the Bộ luật Dân sự 2015 (91/2015/QH13), Luật Thương mại 2005 (36/2005/QH11) and Luật Giao dịch điện tử 2023 (20/2023/QH15) and has not been reviewed by a lawyer. Have it checked against your contracts (form, mandatory content, who may sign, use of company seal) before relying on it.

### Bên A countersignature

With `COMPANY_SIGN_KEY_FILE` and `COMPANY_SIGN_CERT_FILE` set (chart: `company.signKeyPEM` and `company.signCertPEM`, stored in a Secret and mounted read-only), the service signs as Bên A ("chữ ký số của tổ chức") right after it has verified Bên B's signature, and only in PAdES mode, since a detached signature has no document to countersign:

- The countersignature is appended to Bên B's signed PDF as a new revision, so Bên B's signature stays valid and the file carries both, Bên B first. It is time-stamped when `SIGNATURE_TSA_URL` is set.
- The service then verifies its own result like any counterparty would: the signature covers the whole file, the company certificate chains to `SIGNATURE_TRUST_ROOTS_FILE`, is not revoked and is within its validity. If that fails (expired or untrusted company certificate, revoked, TSA down) the renewal is refused with `502` and the contract stays `expired`; nothing is stored.
- The file kept in MinIO (or the database) is the countersigned one; its SHA-256 is that of the final file. `countersigner_subject`, `countersigner_serial`, `countersigner_fingerprint` and `countersigned_at` (the time-stamp's time when there is one) are recorded on the contract next to Bên B's `signer_*` fields.
- The document's Điều 3.2 then says the Phụ lục takes effect when Bên A countersigns; without a company key it says it takes effect when Bên B signs and Bên A records it in the system, as before.

The key is a PEM file read at startup (PKCS#8, PKCS#1 or SEC1; RSA or ECDSA), without a passphrase. The service refuses to start if the key does not match the certificate, and logs a warning when the certificate expires within 30 days. PKCS#11 / HSM tokens and remote signing services are not supported: the private key lives in the pod's Secret, so protect it accordingly (a CA-issued organisational certificate on an HSM would need a different signer). The company is the same for every contract: there is no per-contract or per-user Bên A signer.

## Digital signatures

`POST /contracts/:id/sign` takes `signed_by`, `end_date` (RFC 3339, future), `signing_time` (within 5 minutes of now) and one of two proofs:

- **Detached**: `algorithm`, `signature` (base64) and `certificate` (PEM, leaf first) over the payload from `signing-payload`: a fixed-order JSON document binding contract id, account id, signer, new end date and signing time.
- **PAdES**: `signed_pdf` (base64), the PDF from `document` signed with the signer's own tool (Adobe, a USB token, pyHanko, ...). The service regenerates the document and requires the signed file to start with exactly those bytes, so the signature has to be an appended (incremental) update.

Both are checked against the signer's X.509 certificate (ISO/IEC 9594-8):

- **Algorithms**: `ECDSA-SHA256/384/512` and `RSA-PSS-SHA256/384` (ISO/IEC 14888-3 and -2), plus `RSA-PKCS1-SHA256/384` for legacy CA tokens; PDF signatures may use SHA-256/384/512 digests. EC keys must be at least 256 bits, RSA keys 2048.
- **Trust**: the certificate must allow digital signature or non-repudiation, be valid at the signing time and chain to a CA in `SIGNATURE_TRUST_ROOTS_FILE` (intermediates may follow the leaf, or travel inside the PDF). With no roots configured every signature is refused; `SIGNATURE_ALLOW_UNTRUSTED=true` accepts self-signed certificates for development.
- **Revocation** (`SIGNATURE_REVOCATION`): every certificate on the chain below the root is checked over OCSP, then the CRL named in the certificate. `hard` (default) requires a positive answer, so a certificate without OCSP/CRL info, or a CA whose endpoints are down, is refused; `soft` refuses only on a "revoked" answer and records `revocation_status: unchecked` otherwise; `off` skips it. Responses must be signed by the issuer and current. The URLs are read from certificates that already chain to a trusted root.
- **Time-stamp** (RFC 3161): with `SIGNATURE_TSA_URL` set, every detached signature is time-stamped and the renewal is refused if the authority fails or answers wrongly (wrong imprint or nonce, time more than 5 minutes off, no time-stamping EKU, untrusted or revoked authority). A time-stamp embedded in a PDF signature is verified the same way and, when present, is the time the signer's certificate is judged at. The authority's CA comes from `SIGNATURE_TSA_ROOTS_FILE`, falling back to the trusted roots.
- **PDF signatures** must cover the whole file apart from their own signature bytes, so nothing can be appended or swapped after signing. Any `adbe.pkcs7.detached` signature that passes the checks above is accepted; ETSI PAdES baseline conformance (signing-certificate attributes, LTV data) is not checked.

Stored on the contract and preserved across `PUT`: `signature_format` (`detached-x509` or `pades`), `signature`, `signature_algorithm`, `signer_certificate`, `signer_subject`, `signer_serial`, `signer_fingerprint`, `signature_payload_hash` (of the payload, or of the signed PDF), `revocation_status`, `timestamp_token`, `timestamp_at`, `timestamp_authority`. The signed (and countersigned) PDF is kept as a contract file of kind `renewal`; see "Contract files" for how files are stored.

Bucket versioning or object lock, encryption at rest and per-service MinIO credentials are not set up (the shared MinIO uses one root user, as for the other services); enable versioning/object lock on the bucket if you need write-once evidence storage.

Not covered: long-term validation (embedding OCSP/CRL responses and time-stamps for later re-verification) and hardware-token signing on the server. Whether a signature carries legal weight depends on which CAs you trust and on the applicable law; that has not been reviewed.

Status/stage fields are service-managed: create and `PUT` ignore client values.

If `NOTIFICATION_SERVICE_URL` is set, notification_service is called when a
lead is created or converted, a case is opened or closed, an opportunity is
won, or a contract is approved. Failures are logged and never fail the request.
