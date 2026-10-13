# blockchain-service

Smart contracts for the store (NFT, marketplace, payment escrow) and an HTTP API that drives them through [Starton](https://starton.com), tracks every transaction and publishes status changes to Kafka.

## Contracts

| Contract | Purpose |
| --- | --- |
| `StoreNFT` | ERC721 with per-token URI, ERC2981 royalties, `MINTER_ROLE`, `mint(address,string)` compatible with Starton templates |
| `NFTMarketplace` | Non-custodial fixed-price listings paid in native coin or any ERC20; platform fee (max 10%) and ERC2981 royalties are split on sale; `buy` takes the expected price to protect against front-running price changes |
| `PaymentEscrow` | Holds a buyer's payment for an order (`bytes32` order id) until the buyer or an arbiter releases it to the seller; the seller or an arbiter can refund, and the buyer can refund after the deadline |

```shell
npm install
npm run compile
npm run test:contracts
npm run local-testnet
npm run deploy:local
DEPLOYER_PRIVATE_KEY=... npm run deploy:fuji
```

`scripts/deploy.ts` reads `ADMIN_ADDRESS`, `FEE_RECIPIENT`, `MARKETPLACE_FEE_BPS` (default 250), `ROYALTY_BPS` (default 500), `NFT_NAME`, `NFT_SYMBOL`, `NFT_BASE_URI` (default `ipfs://`), `NFT_CONTRACT_URI`.

## HTTP API

```shell
cp .env.example .env
npm run dev
npm run test:api
```

Without `DATABASE_URL` transactions are kept in memory; without `KAFKA_BROKERS` events are only logged. The table `blockchain_transactions` is created on startup.

### Auth

Same convention as the other services: `Authorization: Bearer <session>` is validated against user-service, `X-Internal-Token` is checked against `INTERNAL_API_TOKENS`. When `INTERNAL_API_TOKENS` is empty, anonymous requests are accepted as internal calls. Endpoints that spend gas from the Starton signer wallet need an admin user or an internal token.

### Endpoints (`/api/v1`)

| Method | Path | Access | Description |
| --- | --- | --- | --- |
| GET | `/health` | public | Liveness |
| POST | `/ipfs/files` | admin | Multipart `file` → IPFS CID |
| POST | `/ipfs/json` | admin | `{name, content}` → IPFS CID |
| POST | `/nfts/metadata` | admin | Multipart `file`, `name`, `description`, `attributes` (JSON array) → uploads image and ERC721 metadata, returns `tokenUri` |
| POST | `/nfts/mint` | admin | `{contractAddress, to, tokenUri}` or `{contractAddress, to, metadata: {name, description, image, attributes}}` |
| POST | `/contracts/deploy` | admin | `{contract: StoreNFT\|NFTMarketplace\|PaymentEscrow, params}` deploys the compiled bytecode |
| POST | `/contracts/deploy-template` | admin | `{templateId, name, params}` deploys a Starton template |
| POST | `/contracts/import` | admin | `{contract, address}` registers a contract deployed with Hardhat in Starton |
| POST | `/contracts/:network/:address/call` | admin | `{functionName, params, value?}` write call, tracked |
| POST | `/contracts/:network/:address/read` | user | `{functionName, params}` |
| GET | `/marketplace/:network/:address/listings/:id` | user | Listing details |
| GET | `/escrow/:network/:address/orders/:orderId` | user | Escrow payment and status |
| POST | `/escrow/:network/:address/orders/:orderId/release` | admin | Arbiter releases the payment to the seller |
| POST | `/escrow/:network/:address/orders/:orderId/refund` | admin | Arbiter refunds the buyer |
| GET | `/transactions` | user | `status`, `contractAddress`, `limit`, `offset`; users only see their own, admins can filter by `requestedBy` |
| GET | `/transactions/:id` | owner/admin | Tracked transaction |
| POST | `/transactions/:id/sync` | admin | Refresh status from Starton |
| POST | `/webhooks/starton` | signature | Starton webhook receiver |

`network` defaults to `STARTON_NETWORK`, `signerWallet` to `WALLET_ADDRESS`. Escrow order ids that are not already a 32-byte hex string are hashed with `keccak256`, the same as `ethers.utils.id(orderId)` on the client. `tokenUri` is appended to the contract base URI, so with the default `ipfs://` base pass the bare CID.

Calls and reads only work on contracts Starton knows: deploy through `/contracts/deploy` or register with `/contracts/import`. Users buy, list and fund escrow from their own wallets, so those calls go directly from the frontend to the contracts.

### Webhooks and events

Create a Starton webhook for transaction events pointing to `/api/v1/webhooks/starton` and put its secret in `STARTON_WEBHOOK_SECRET`. The `starton-signature` header (HMAC-SHA256 of the raw body) is verified; requests are rejected while the secret is unset. Final statuses (`SUCCESS`, `ERROR`, `DROPPED`) are never overwritten by late events. Transactions not created through this API are stored as `external`.

Each created or changed transaction is published to `KAFKA_TOPIC` (default `blockchain.transactions`) with the transaction id as key:

```json
{"type": "transaction.updated", "occurredAt": "...", "data": {"id": "...", "kind": "mint", "status": "SUCCESS", "transactionHash": "0x...", "...": "..."}}
```

## Starton playground

`npm run starton:demo` runs the original flow: deploy an ERC721 template, upload `scripts/assets/img.png` and mint it to `WALLET_ADDRESS`.
