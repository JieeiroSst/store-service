# filter-service

GraphQL gateway over the GET APIs of the other services in this repo:
84 services, 687 GET endpoints, one typed query field each.

```graphql
{
  webrtcService { room(room_id: "lobby") { capacity participants { user_id } } }
  bookService { books(limit: 5) { data { title author } pagination { nextCursor } } }
  services { name envVar baseURL endpoints }
}
```

Each service is a namespace field on `Query` (`bookService`, `ticketService`, …)
whose fields map 1:1 to that service's GET endpoints. The resolver calls the
service over HTTP and decodes the JSON into the generated types; a failing
service only nulls its own fields and adds a GraphQL error.

Playground: `GET /`, endpoint: `POST|GET /query`, liveness: `GET /health`.

## How it is built

```
spec/<service>.yaml          what to expose: endpoints + GraphQL types (hand-maintained)
tools/gen                    spec -> schema, typed clients, resolvers
gql/schema/*.graphqls        generated per service (+ schema.graphqls, the root)
adapter/<service>/           generated typed HTTP client per service
adapter/clients_gen.go       generated registry, base URLs from env
adapter/rest                 shared GET client: header forwarding, envelopes, errors
gql/model, gql/generated     gqlgen output
gql/resolver                 generated resolvers + resolver.go
```

Regenerate after editing a spec: `make gen` (runs `tools/gen`, then
`go tool gqlgen generate`). `make build`, `make vet`, `make test`;
`TestEveryEndpointIsWired` calls every query field against a stub upstream. `make check-gen` fails when the committed generated code is out of
date with `spec/`.

CI/CD: the `filter_service` job in `.github/workflows/go.yml` runs build, vet,
test and check-gen on every push; `filter-service-cicd.yml` (manual) also
publishes `ghcr.io/jieeirosst/filter-service`. The chart is
`chart/filter-service`, deployed by Argo CD like the other services.

### Spec format

```yaml
service: webrtc-service          # repo directory / service name
source: webrtc-service/...       # where the routes are defined (for humans)
namespace: webrtcService         # field on Query
prefix: Webrtc                   # every type name starts with it
baseURL: http://webrtc-service-svc
unwrap: data                     # optional default envelope key (dot path)
forwardHeaders: [X-Api-Key]      # optional extra incoming headers to forward
endpoints:
  - field: room
    path: /api/rooms/{room_id}   # {x} = path argument (String! unless in params)
    params: { room_id: String! }
    query: { limit: Int }        # query-string arguments
    headers: { X-User-Id: String! }  # header arguments
    returns: WebrtcRoomInfo      # GraphQL type
    unwrap: ""                   # per-endpoint envelope override
    rawParams: [path]            # wildcard params keep their slashes
    doc: shown in the schema
types: |
  type WebrtcRoomInfo { room_id: String capacity: Int ... }
```

Conventions:

- GraphQL field names are the services' JSON keys, so responses decode
  without mapping. Keys that are not valid GraphQL names use
  `@goTag(key: "json", value: "<original-key>")`.
- Fields and results are nullable and list items non-null, so a missing or
  null value from a service never fails the whole query.
- grpc-gateway services (account-transaction, authorize, car-rental,
  hospital, user) use protojson: camelCase keys, 64-bit integers as strings.

### Headers

`Authorization`, `Cookie`, `X-Request-Id` and `Accept-Language` from the
GraphQL request are forwarded to every service, plus each spec's
`forwardHeaders` to that service only (`X-Api-Key` for manage_service,
`X-API-Key` for notifyhub-service and shipping-service).

## Services and base URLs

Every base URL defaults to the in-cluster Service from `chart/` (or
`http://<service>:<port>` when the service has no chart) and can be
overridden with the variable below.

| Service | Namespace | Endpoints | Env var | Default base URL |
|---|---|---|---|---|
| account-transaction-service | `accountTransactionService` | 5 | `ACCOUNT_TRANSACTION_SERVICE_URL` | http://account-transaction-api-svc |
| address-country-service | `addressCountryService` | 14 | `ADDRESS_COUNTRY_SERVICE_URL` | http://address-country-svc |
| admanagement-service | `admanagementService` | 23 | `ADMANAGEMENT_SERVICE_URL` | http://admanagement-svc |
| ai-agent-system | `aiAgentSystem` | 2 | `AI_AGENT_SYSTEM_URL` | http://ai-agent-system:8080 |
| airflow-service | `airflowService` | 11 | `AIRFLOW_SERVICE_URL` | http://airflow-service-svc |
| arrange-service | `arrangeService` | 4 | `ARRANGE_SERVICE_URL` | http://arrange-service-svc |
| authorize_service | `authorizeService` | 2 | `AUTHORIZE_SERVICE_URL` | http://authorize-service-api-svc |
| automatic-payment-service | `automaticPaymentService` | 6 | `AUTOMATIC_PAYMENT_SERVICE_URL` | http://automatic-payment-service-svc |
| banking-service | `bankingService` | 16 | `BANKING_SERVICE_URL` | http://banking-service-svc |
| basket-service | `basketService` | 9 | `BASKET_SERVICE_URL` | http://basket-service-svc |
| billing-service | `billingService` | 12 | `BILLING_SERVICE_URL` | http://billing-service-svc |
| bonuslink-service | `bonuslinkService` | 2 | `BONUSLINK_SERVICE_URL` | http://bonuslink-service-svc |
| book-service | `bookService` | 7 | `BOOK_SERVICE_URL` | http://book-service-svc |
| book-store-service | `bookStoreService` | 13 | `BOOK_STORE_SERVICE_URL` | http://book-store-service-svc |
| booking-mini-service | `bookingMiniService` | 3 | `BOOKING_MINI_SERVICE_URL` | http://booking-mini-service-svc |
| bot-service | `botService` | 2 | `BOT_SERVICE_URL` | http://bot-service-svc |
| calculate-service | `calculateService` | 6 | `CALCULATE_SERVICE_URL` | http://calculate-service-svc |
| call_center_ai | `callCenterAi` | 6 | `CALL_CENTER_AI_URL` | http://call-center-ai:8000 |
| car-rental-service | `carRentalService` | 10 | `CAR_RENTAL_SERVICE_URL` | http://car-rental-service-svc |
| catalogues-service | `cataloguesService` | 8 | `CATALOGUES_SERVICE_URL` | http://catalogues-service-svc |
| cdn-service | `cdnService` | 2 | `CDN_SERVICE_URL` | http://cdn-service-svc |
| chat_service | `chatService` | 2 | `CHAT_SERVICE_URL` | http://chat-service:6060 |
| chatbot-system | `chatbotSystem` | 2 | `CHATBOT_SYSTEM_URL` | http://chatbot-system:8080 |
| coupon-service | `couponService` | 8 | `COUPON_SERVICE_URL` | http://coupon-service-svc |
| customer-relationship-service | `customerRelationshipService` | 28 | `CUSTOMER_RELATIONSHIP_SERVICE_URL` | http://customer-relationship-service-svc |
| doordash-service | `doordashService` | 4 | `DOORDASH_SERVICE_URL` | http://doordash-service-svc |
| draw-image-service | `drawImageService` | 1 | `DRAW_IMAGE_SERVICE_URL` | http://draw-image-service-svc |
| ekyc-service | `ekycService` | 1 | `EKYC_SERVICE_URL` | http://ekyc-service-svc |
| face-recognition-service | `faceRecognitionService` | 1 | `FACE_RECOGNITION_SERVICE_URL` | http://face-recognition-service:1234 |
| food-service | `foodService` | 3 | `FOOD_SERVICE_URL` | http://food-service:8080 |
| geoservice | `geoservice` | 9 | `GEOSERVICE_URL` | http://geoservice-svc |
| hospital-patient-management-service | `hospitalPatientManagementService` | 21 | `HOSPITAL_PATIENT_MANAGEMENT_SERVICE_URL` | http://hospital-patient-management-service-svc |
| identified_service | `identifiedService` | 3 | `IDENTIFIED_SERVICE_URL` | http://identified-service:8000 |
| integrated-payment-service | `integratedPaymentService` | 2 | `INTEGRATED_PAYMENT_SERVICE_URL` | http://integrated-payment-service:8080 |
| kms-service | `kmsService` | 6 | `KMS_SERVICE_URL` | http://kms-service:8080 |
| livestream_service | `livestreamService` | 6 | `LIVESTREAM_SERVICE_URL` | http://livestream-service-svc |
| lottery-service | `lotteryService` | 1 | `LOTTERY_SERVICE_URL` | http://lottery-service-svc |
| manage_service | `manageService` | 89 | `MANAGE_SERVICE_URL` | http://manage-service-svc |
| media-service | `mediaService` | 1 | `MEDIA_SERVICE_URL` | http://media-service:8080 |
| medical-service | `medicalService` | 10 | `MEDICAL_SERVICE_URL` | http://medical-service-svc |
| movie-recommendation-service | `movieRecommendationService` | 7 | `MOVIE_RECOMMENDATION_SERVICE_URL` | http://movie-recommendation-service-svc |
| notification_service | `notificationService` | 9 | `NOTIFICATION_SERVICE_URL` | http://notification-service-svc |
| notifyhub-service | `notifyhubService` | 10 | `NOTIFYHUB_SERVICE_URL` | http://notifyhub-service:8095 |
| ollama-model-py | `ollamaModelPy` | 1 | `OLLAMA_MODEL_PY_URL` | http://ollama-model-py:8000 |
| ollama-service | `ollamaService` | 5 | `OLLAMA_SERVICE_URL` | http://ollama-service:8080 |
| parking-lot-service | `parkingLotService` | 4 | `PARKING_LOT_SERVICE_URL` | http://parking-lot-service-svc |
| partner-service | `partnerService` | 11 | `PARTNER_SERVICE_URL` | http://partner-service-svc |
| payer-service | `payerService` | 1 | `PAYER_SERVICE_URL` | http://payer-service:8080 |
| payment-wallet-service | `paymentWalletService` | 11 | `PAYMENT_WALLET_SERVICE_URL` | http://payment-wallet-service-svc |
| payment_service | `paymentService` | 3 | `PAYMENT_SERVICE_URL` | http://payment-service-svc |
| photo-service | `photoService` | 1 | `PHOTO_SERVICE_URL` | http://photo-service-svc |
| point-service | `pointService` | 6 | `POINT_SERVICE_URL` | http://point-service:3000 |
| polymarket-service | `polymarketService` | 26 | `POLYMARKET_SERVICE_URL` | http://polymarket-service-svc |
| post_service | `postService` | 4 | `POST_SERVICE_URL` | http://post-service-svc |
| qr-service | `qrService` | 4 | `QR_SERVICE_URL` | http://qr-service:8080 |
| recompense-service | `recompenseService` | 5 | `RECOMPENSE_SERVICE_URL` | http://recompense-service-svc |
| recruitment-platform-service | `recruitmentPlatformService` | 16 | `RECRUITMENT_PLATFORM_SERVICE_URL` | http://recruitment-platform-service:8080 |
| referral-service | `referralService` | 6 | `REFERRAL_SERVICE_URL` | http://referral-service-svc |
| rent_house_service | `rentHouseService` | 17 | `RENT_HOUSE_SERVICE_URL` | http://rent-house-service-svc |
| reservation-service | `reservationService` | 23 | `RESERVATION_SERVICE_URL` | http://reservation-service-svc |
| restaurant-accounting-service | `restaurantAccountingService` | 1 | `RESTAURANT_ACCOUNTING_SERVICE_URL` | http://accounting-service-svc |
| restaurant-consumer-service | `restaurantConsumerService` | 1 | `RESTAURANT_CONSUMER_SERVICE_URL` | http://consumer-service-svc |
| restaurant-delivery-service | `restaurantDeliveryService` | 1 | `RESTAURANT_DELIVERY_SERVICE_URL` | http://delivery-service-svc |
| restaurant-kitchen-service | `restaurantKitchenService` | 4 | `RESTAURANT_KITCHEN_SERVICE_URL` | http://kitchen-service-svc |
| restaurant-order-service | `restaurantOrderService` | 4 | `RESTAURANT_ORDER_SERVICE_URL` | http://order-service-svc |
| reward-service | `rewardService` | 2 | `REWARD_SERVICE_URL` | http://reward-service:8080 |
| room-service | `roomService` | 5 | `ROOM_SERVICE_URL` | http://room-service-svc |
| sard-service | `sardService` | 1 | `SARD_SERVICE_URL` | http://sard-customer-info-service:8081 |
| shipping-service | `shippingService` | 7 | `SHIPPING_SERVICE_URL` | http://shipping-service-svc |
| shopify-service | `shopifyService` | 4 | `SHOPIFY_SERVICE_URL` | http://shopify-service-svc |
| shortlink-service | `shortlinkService` | 14 | `SHORTLINK_SERVICE_URL` | http://shortlink-service-svc |
| solana-service | `solanaService` | 15 | `SOLANA_SERVICE_URL` | http://solana-service:8080 |
| threads-service | `threadsService` | 8 | `THREADS_SERVICE_URL` | http://threads-service-svc |
| ticket-service | `ticketService` | 34 | `TICKET_SERVICE_URL` | http://ticket-service-svc |
| toggle-service | `toggleService` | 11 | `TOGGLE_SERVICE_URL` | http://toggle-service-svc |
| tool-service | `toolService` | 7 | `TOOL_SERVICE_URL` | http://tool-service-svc |
| upload_service | `uploadService` | 2 | `UPLOAD_SERVICE_URL` | http://upload-service-svc |
| user_service | `userService` | 3 | `USER_SERVICE_URL` | http://user-api-svc |
| vending-machine-service | `vendingMachineService` | 14 | `VENDING_MACHINE_SERVICE_URL` | http://vending-machine-service-svc |
| video-service | `videoService` | 3 | `VIDEO_SERVICE_URL` | http://video-service-svc |
| voucher-service | `voucherService` | 9 | `VOUCHER_SERVICE_URL` | http://voucher-service-svc |
| wallet-service | `walletService` | 2 | `WALLET_SERVICE_URL` | http://wallet-service:8080 |
| webrtc-service | `webrtcService` | 2 | `WEBRTC_SERVICE_URL` | http://webrtc-service-svc |
| wishlists-service | `wishlistsService` | 2 | `WISHLISTS_SERVICE_URL` | http://wishlists-service:8000 |

## Not exposed

GET routes that are not JSON APIs: `/health`-style probes, `/metrics`,
swagger, WebSocket upgrades, SSE streams, redirects, HTML pages and file
downloads (images, PDF, CSV, HLS, QR codes, artifacts).

Others, on purpose:

| Route | Why |
|---|---|
| ticket-service `/internal/v1/*` | service-to-service API keyed by a shared secret that acts as admin |
| manage_service `GET /token`, `GET /client` | read a JSON body, which a GET from the gateway cannot send |
| face-service, networking-service, sso-service | HTML pages / redirects / text only |
| message-service, offer-range-service, subsidies-service, search-service, real-time-service | no GET routes registered (or WebSocket only) |
| gateway-api, gateway-internal-api | GraphQL gateways themselves |

Fields a service returns but the gateway leaves out of the schema:
`password` (user_service users), `config` and `auth_config`
(notifyhub-service channels and data sources, provider credentials), and the
webhook signing `secret` (shortlink-service).
