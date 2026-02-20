# QR Service

Dynamic QR Code Generator API - Go + Gin + MongoDB + Uber FX + Hexagonal Architecture

## Architecture

```
qr-service/
├── cmd/
│   └── main.go                          # Entry point + FX bootstrap
├── internal/
│   ├── domain/                          # Pure business logic (no deps)
│   │   ├── entity/qr.go                 # QRCode & ScanHistory entities
│   │   └── port/
│   │       ├── repository.go            # Driven ports (storage interfaces)
│   │       └── service.go               # Driving ports (use case interfaces)
│   ├── application/service/
│   │   ├── qr_service.go                # QR code use cases
│   │   └── scan_history_service.go      # Scan history use cases
│   ├── infrastructure/
│   │   ├── config/config.go             # Viper config loader
│   │   ├── repository/
│   │   │   ├── mongo.go                 # MongoDB connection
│   │   │   ├── qr_repository.go         # MongoDB QR adapter
│   │   │   └── scan_history_repository.go
│   │   └── http/
│   │       ├── handler/qr_handler.go    # Gin HTTP handlers
│   │       └── router/router.go         # Route definitions
│   └── module/module.go                 # Uber FX module wiring
├── pkg/
│   ├── logger/logger.go
│   ├── middleware/middleware.go
│   └── response/response.go
├── config.yaml
├── docker-compose.yml
└── Dockerfile
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/qr | Generate new QR code |
| GET | /api/v1/qr | List (paginated + filters) |
| GET | /api/v1/qr/:id | Get by ID |
| PUT | /api/v1/qr/:id | Update metadata |
| PATCH | /api/v1/qr/:id/content | **Dynamic update** target URL |
| DELETE | /api/v1/qr/:id | Delete + history |
| POST | /api/v1/qr/:id/regenerate | Regenerate image |
| GET | /qr/scan/:shortCode | Scan redirect (logs history) |
| GET | /api/v1/qr/:id/history | Scan history |
| GET | /api/v1/qr/:id/stats | Analytics & device breakdown |

## Quick Start

```bash
# Docker Compose (includes MongoDB)
docker-compose up -d

# OR local dev
go mod tidy
go run ./cmd/main.go
```

## Example: Generate Dynamic QR

```bash
# Create
curl -X POST http://localhost:8080/api/v1/qr \
  -H "Content-Type: application/json" \
  -d '{"title":"My Site","type":"url","content":"https://example.com","redirect_url":"https://example.com"}'

# Update target (QR image stays the same!)
curl -X PATCH http://localhost:8080/api/v1/qr/{id}/content \
  -H "Content-Type: application/json" \
  -d '{"content":"https://new-url.com","redirect_url":"https://new-url.com"}'
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| APP_APP_BASE_URL | http://localhost:8080 | Base URL for scan redirects |
| APP_MONGODB_URI | mongodb://localhost:27017 | MongoDB URI |
| APP_MONGODB_DATABASE | qr_service | Database name |
| APP_SERVER_PORT | 8080 | Server port |
