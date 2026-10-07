# sard-auth-service

Payment authentication for sARD cards, in the style of 3-D Secure. Before a
merchant authorizes a high-value online payment, the cardholder confirms it
with a one-time password (OTP) while logged in to `user-service`.

## Flow

```
merchant/gateway          sard-auth-service              cardholder app (user-service session)     sard-card-service
      │ POST /api/v1/payment-authentications ─►│ resolve PAN ──────────────────────────────────────────►│ /internal/v1/cards/resolve
      │                   │ CreateOTP{username} ──► authorize-service (gRPC); OTP sent to the card owner
      │                   │◄── GET  /api/v1/me/payment-authentications         (Bearer <session token>)
      │                   │◄── POST /api/v1/me/payment-authentications/{id}/verify {"otp"}
      │ POST /api/v1/authorizations {"authentication_id": id} ─────────────────────────────────────────────►│
      │                   │◄──────────── POST /internal/v1/payment-authentications/{id}/consume ──────────────│
```

1. The merchant creates a challenge with the PAN, expiry, amount, currency and
   merchant. The service asks `sard-card-service` who owns the card. The card
   must be NORMAL, and its account must have an enrolled `user_id`.
2. The service reads the cardholder's `username` from user-service
   (`GET /internal/v1/users/{id}`). It then asks **authorize-service** for an
   OTP over gRPC: `CreateOTP{username}` returns `{otp, expires_at}`. The code is
   a 6-digit TOTP that lives 30-60 seconds, and authorize-service allows at
   most 5 codes per username per 24h. This service never generates or stores
   OTPs itself. The code is delivered through the `otpDelivery` backend: `log`
   for development, or `webhook` to POST it to a notification service.
3. The cardholder confirms with their **user-service session**. The service
   validates it with `POST /api/v1/validate` on user-service, and only the
   card's owner can verify, resend or decline the challenge.
   - The code is checked by authorize-service (`AuthorizeOTP{username, otp}`
     returns `valid`).
   - There are 3 wrong attempts, then the challenge is FAILED.
   - Once the code expires, verify returns `410` without counting an attempt,
     and the cardholder calls `.../resend` (at most `maxResends` times) for a
     new code.
   - The challenge itself expires after `otpTtl` (5 minutes).
   - An authorize-service outage returns `503` without counting an attempt.
   - Verify, resend and decline are rate limited per session.
4. A successful verification sets AUTHENTICATED for `validityTime`
   (10 minutes). `sard-card-service` then consumes it exactly once. The card
   and currency must match and the amount must not exceed the authenticated
   amount; after that the challenge is USED.

Statuses: PENDING, AUTHENTICATED, USED, FAILED, DECLINED and EXPIRED.

## API

| Zone | Auth | Routes |
|---|---|---|
| Merchant/gateway | `ACCESS_TOKENS` | `POST /api/v1/payment-authentications`, `GET /api/v1/payment-authentications/{id}` |
| Cardholder | user-service session (`Authorization: Bearer` or `X-Session-Token`) | `GET /api/v1/me/payment-authentications`, `POST /api/v1/me/payment-authentications/{id}/verify` `{"otp"}`, `POST .../{id}/resend`, `POST .../{id}/decline` |
| Internal | `INTERNAL_TOKENS` | `POST /internal/v1/payment-authentications/{id}/consume` `{"card_id","amount","currency"}` |

Errors: `401` bad token or session, `403` someone else's challenge, `404`,
`409` wrong state or mismatch, `410` challenge or OTP expired, `422` wrong OTP
(the response includes `attempts_left`), `429` rate limited or authorize-service's
OTP limit reached, `503` a dependency is down.

## Architecture

Hexagonal (ports and adapters), wired with fx: `domain` (Challenge state
machine, Policy) → `port` → `application` (AuthenticationService) → `adapter`.
The adapters are `http` (with security headers and a rate limiter),
`postgres`/`memory`, the `authorizeservice` gRPC client (OTP), the
`cardservice` and `userservice` clients, `otp` (log and webhook senders),
metrics and clock.

## Configuration

The Consul key is `auth_service`, and environment variables override it:
- `AUTHORIZE_SERVICE_GRPC_ADDR` (default in Consul: `authorize-service-grpc-svc:90`) and `AUTHORIZE_SERVICE_TIMEOUT`
- `ACCESS_TOKENS` and `INTERNAL_TOKENS`
- `USER_SERVICE_URL` and `USER_SERVICE_TOKEN`
- `CARD_SERVICE_URL` and `CARD_SERVICE_TOKEN`
- `OTP_TTL`, `OTP_MAX_ATTEMPTS`, `OTP_MAX_RESENDS` and `AUTHENTICATION_VALIDITY`
- `OTP_DELIVERY` (`log` or `webhook`), `OTP_WEBHOOK_URL` and `OTP_WEBHOOK_TOKEN`
- `VERIFY_PER_MINUTE`
- `EXPOSE_OTP=true` returns the OTP in the create response (tests only)
- `STORAGE_BACKEND`, `POSTGRES_*` and `PORT_HTTP_SERVER`

The PAN tokenization vault in the old `main.go` was removed:
`sard-card-service` already encrypts PANs and keeps their HMACs.
