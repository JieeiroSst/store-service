# recompense-service

To install dependencies:

```bash
bun install
```

To run:

```bash
bun run server.ts
```

This project was created using `bun init` in bun v1.1.3. [Bun](https://bun.sh) is a fast all-in-one JavaScript runtime.

## Architecture (hexagonal / ports & adapters)

```
server.ts                      entry point
config/                        env config
internal/
  domain/                      entities + rules, no I/O
  application/
    port/inbound/              what the app offers (RecompenseService)
    port/outbound/             what the app needs (repository, cache, id generator)
    service/                   use cases, implement inbound ports using outbound ports
  adapter/
    inbound/http/              Bun HTTP handler + request DTO
    outbound/postgres/         repository + migration
    outbound/redis/            read cache (+ no-op fallback)
    outbound/snowflake/        id generator
  bootstrap/container.ts       composition root (wires adapters to ports)
```

Dependencies point inward: adapters -> application -> domain. Only `bootstrap` knows concrete adapters.

## Loyalty features

Points are stored in an append-only ledger (`points_ledger`); tier is derived from the balance.

| Endpoint | Purpose |
| --- | --- |
| `POST /api/members/:id/points/earn` `{amount, idempotencyKey}` | Credit points for a purchase. Replaying a key credits nothing. |
| `GET /api/members/:id/points` | Balance, tier, points to next tier. |
| `GET /api/members/:id/recompenses/best?amount=` | Most valuable usable recompense for a purchase. |
| `POST /api/recompenses/:id/redeem` `{memberId, amount}` | Price a purchase with one recompense (422 if not usable). |

Algorithms (pure, in `internal/domain/loyalty.ts`, tested in `tests/`):
- **Tiers**: bronze 0, silver 1000, gold 5000, platinum 20000 points; earn multiplier 1 / 1.25 / 1.5 / 2.
- **Earning**: `floor(floor(amount / 10) * multiplier)`, using the tier held before the purchase.
- **Discount**: `fixed` or `percent` (with `MaxReduction`), never above the purchase.
- **Best pick**: among non-expired recompenses the member's tier and points allow, the largest discount wins; ties go to the earliest expiry, then lowest id.
