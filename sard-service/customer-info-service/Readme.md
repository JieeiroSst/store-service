# sard-customer-info-service

Customer records for sARD. A customer is created from a `user-service` user,
and its profile (name, email, phone, address, gender, locked or active) is
kept in sync with that user. The service decides KYC from the results of
**ekyc-service**, the store service that owns CCCD reading (MRZ OCR, NFC chip)
and face biometrics. `sard-card-service` only opens card accounts for
customers whose `card_eligibility.eligible` is `true`.

```
app ──► customer-info: POST /customers/{id}/kyc (front, back) ──► ekyc-service: POST /api/v1/ekyc/{user_id}/citizen-card  (service token)
app ──► ekyc-service:  POST /api/v1/ekyc/{user_id}/face-scan, /verify, /nfc-chip                                      (user-service session)
app ──► customer-info: POST /customers/{id}/kyc/refresh       ──► ekyc-service: GET /api/v1/ekyc/{user_id}               (service token)
```

The app calls ekyc-service directly with the user's own user-service session,
and ekyc-service only accepts a session for that same user. This service calls
it with `EKYC_SERVICE_TOKEN`, which must be listed in ekyc-service's
`INTERNAL_TOKENS`.

## Business rules

- **Onboarding.** `POST /api/v1/customers {"user_id"}` reads the user from
  `user-service` (`GET /internal/v1/users/{id}`). The call is idempotent: the
  same user always maps to the same customer, even under concurrent calls. A
  user that is locked in user-service cannot be onboarded.
- **Sync.** `POST /api/v1/customers/{id}/sync` re-reads the profile.
  - If the user is locked in user-service, the customer becomes SUSPENDED.
  - If the user is unlocked again, the customer becomes ACTIVE.
  - If the name changes and no longer matches a verified document, KYC goes
    to REVIEW_REQUIRED.
- **KYC.** The decision is made by this service, from what ekyc-service has on
  file for the user (keyed by the user-service id):
  - The document number is the 12-digit CCCD number. It is taken from the
    optional field of MRZ line 1, because the MRZ document field only holds
    its last 9 digits.
  - The MRZ check digits must be valid, unless the data came from a verified
    NFC chip.
  - The surname and given names must match the user-service name, ignoring
    diacritics.
  - Dates are read from the MRZ (YYMMDD). The customer must be at least
    `kyc.minAge` (18), and the document must not be expired.
  - The OCR confidence must be at least `kyc.minConfidence`, unless the data
    came from the NFC chip.
  - When `kyc.requireFaceMatch` is set (the default), ekyc-service's face
    verification must be `verified`. Until then KYC is **PENDING**.
    ekyc-service resets its face verification whenever a new card or chip is
    submitted.

  The result is NONE, PENDING, VERIFIED or REJECTED (with reasons). A CCCD can
  be verified for only one customer. Only an HMAC of the number and its last 4
  digits are stored.
- **Card eligibility.** The customer is ACTIVE, and KYC is VERIFIED with a
  document that has not expired.

## API

`/api/*` routes need `Authorization: Bearer <token>` when `ACCESS_TOKENS` is set.

| Method | Path | Body |
|---|---|---|
| POST | `/api/v1/customers` | `{"user_id"}` returns `201` when created, `200` when the customer already exists |
| GET  | `/api/v1/customers/{id}` · `/api/v1/users/{userId}/customer` | |
| POST | `/api/v1/customers/{id}/sync` | |
| POST | `/api/v1/customers/{id}/kyc` | multipart `front` and `back` (each at most 10 MB), forwarded to ekyc-service |
| POST | `/api/v1/customers/{id}/kyc/refresh` | re-evaluates from ekyc-service, after a face scan or NFC read |

## Architecture

Hexagonal (ports and adapters), wired with fx: `domain` (Customer, KYC
policy) → `port` → `application` (CustomerService) → `adapter`. The adapters
are `http`, `postgres`/`memory`, `userservice` and `ekyc` (HTTP clients),
`crypto` (document HMAC), metrics and clock.

## Configuration

The Consul key is `customer_info_service`, and environment variables override
it:
- `DOCUMENT_HASH_KEY` (required, at least 32 characters)
- `ACCESS_TOKENS`
- `USER_SERVICE_URL` and `USER_SERVICE_TOKEN`
- `EKYC_SERVICE_URL` (default in Consul: `http://ekyc-service-svc`), `EKYC_SERVICE_TOKEN` and `EKYC_SERVICE_TIMEOUT`
- `KYC_MIN_AGE`, `KYC_MIN_CONFIDENCE` and `KYC_REQUIRE_FACE_MATCH`
- `STORAGE_BACKEND`, `POSTGRES_*` and `PORT_HTTP_SERVER`
