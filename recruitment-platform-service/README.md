# Recruitment Platform

Go · Hexagonal Architecture · Temporal · PostgreSQL · pgvector · Uber FX

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Inbound Adapters                                │
│   HTTP / Gin handlers     Temporal Signal receivers     CLI             │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │ calls
┌────────────────────────────────▼────────────────────────────────────────┐
│                       Inbound Ports (interfaces)                        │
│  CandidateService  JobService  ApplicationService  ReferralService      │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │ implements
┌────────────────────────────────▼────────────────────────────────────────┐
│                         Use Cases / Application Layer                   │
│  usecase/candidate  usecase/job  usecase/application  usecase/referral  │
│                   (pure Go, no framework dependency)                    │
└────────────┬──────────────┬───────────────────┬───────────────┬─────────┘
             │              │                   │               │
     ┌───────▼──────┐ ┌─────▼──────┐ ┌─────────▼──────┐ ┌─────▼─────────┐
     │   Domain     │ │  Outbound  │ │  Outbound Port │ │ Outbound Port │
     │   (pure)     │ │  Port      │ │  WorkflowSvc   │ │  AI / Notif.  │
     │  candidate   │ │  Repos     │ │                │ │               │
     │  job         │ └─────┬──────┘ └────────┬───────┘ └──────┬────────┘
     │  application │       │                 │                 │
     │  referral    │       │                 │                 │
     └──────────────┘       │                 │                 │
                            │ implements      │ implements      │
         ┌──────────────────▼──┐  ┌───────────▼─────────┐ ┌───▼──────────┐
         │ Persistence Adapter │  │  Temporal Adapter   │ │ AI Adapter  │
         │  postgres/           │  │  worker.go          │ │  openai.go  │
         │  candidate_repo.go  │  │  workflow/*.go       │ │             │
         │  job_repo.go        │  │  activity/*.go       │ └─────────────┘
         │  application_repo   │  └─────────────────────┘
         └──────────┬──────────┘
                    │
         ┌──────────▼──────────┐
         │   PostgreSQL        │
         │   + pgvector        │
         └─────────────────────┘
```

## Directory Structure

```
recruitment-platform/
├── cmd/server/main.go              # Entry point (FX bootstrap)
├── config/config.yaml
├── migrations/                     # Goose SQL migrations
│   └── 001_init.sql
├── internal/
│   ├── domain/                     # ① Pure domain (no deps)
│   │   ├── candidate/
│   │   │   ├── entity.go           # Entity + value objects + domain logic
│   │   │   └── repository.go       # Outbound port interface
│   │   ├── job/entity.go
│   │   ├── application/entity.go
│   │   ├── referral/entity.go
│   │   └── shared/types.go         # BaseEntity, Money, AIScore, Events
│   │
│   ├── port/ports.go               # ② All inbound + outbound port interfaces
│   │
│   ├── usecase/                    # ③ Application layer (implements inbound ports)
│   │   ├── candidate/service.go
│   │   ├── job/service.go
│   │   ├── application/service.go  # Full recruitment lifecycle orchestration
│   │   └── referral/service.go
│   │
│   ├── adapter/                    # ④ Infrastructure adapters
│   │   ├── persistence/postgres/   # Implements Repository interfaces
│   │   │   ├── candidate_repo.go
│   │   │   ├── job_repo.go
│   │   │   └── application_repo.go
│   │   ├── temporal/
│   │   │   ├── worker.go           # Worker + WorkflowService adapter
│   │   │   ├── workflow/
│   │   │   │   ├── recruitment.go         # Main lifecycle workflow
│   │   │   │   └── candidate_referral.go  # Candidate nurture + referral payout
│   │   │   └── activity/
│   │   │       └── activities.go          # Notify, AI score, payout activities
│   │   ├── http/
│   │   │   └── handler/
│   │   │       ├── application.go
│   │   │       ├── candidate.go
│   │   │       ├── job.go
│   │   │       └── referral.go
│   │   └── ai/                     # AI service adapter (OpenAI / Anthropic)
│   │
│   └── module/                     # ⑤ FX DI modules
│       ├── modules.go              # DatabaseModule, TemporalModule, ServiceModule, HTTPModule
│       └── config.go
│
├── docker-compose.yml
└── Dockerfile
```

## Key Design Decisions

### Hexagonal Architecture
- **Domain** is 100% pure Go with zero external dependencies
- **Ports** are Go interfaces – the domain never knows about postgres, temporal, or HTTP
- **Adapters** live in `internal/adapter/` and wire everything via FX

### Temporal Workflow Strategy
| Workflow | Purpose |
|---|---|
| `RecruitmentLifecycleWorkflow` | Per-application state machine; handles stage signals, interview reminders, offer expiry, SLA timeout |
| `CandidateLifecycleWorkflow` | Long-running; sends nurture emails after 30-day idle, terminates on hire/blacklist |
| `ReferralNetworkWorkflow` | Waits for hire/probation confirmation, computes commission, triggers payout |

### AI Integration
- Resume parsing → skills, experience level, embeddings stored in `vector(1536)` column
- pgvector `ivfflat` index for ANN candidate-job similarity search
- Match scoring is triggered async on application creation

### Referral Network
- Multi-level partner hierarchy (`referred_by_partner_id`)
- Tier auto-upgrade: Bronze → Silver → Gold → Platinum by `hired_referrals`
- Commission: fixed or % of salary, triggered on hire or probation pass
- All payout logic encapsulated in `ReferralNetworkWorkflow`

## Running Locally

```bash
# Start infrastructure
docker compose up postgres temporal temporal-ui -d

# Run migrations
goose -dir migrations postgres "host=localhost user=postgres password=secret dbname=recruitment_db sslmode=disable" up

# Run app
go run ./cmd/server
```

Open Temporal UI at http://localhost:8088

## API Quick Reference

```
POST   /api/v1/candidates
GET    /api/v1/candidates?status=new&skills[]=golang&page=1
GET    /api/v1/candidates/:id

POST   /api/v1/jobs
PUT    /api/v1/jobs/:id/publish

POST   /api/v1/applications                     # Apply
PUT    /api/v1/applications/:id/stage           # Move pipeline stage
POST   /api/v1/applications/:id/interviews      # Schedule interview
POST   /api/v1/applications/:id/interviews/:ivId/feedback
POST   /api/v1/applications/:id/offer           # Extend offer
POST   /api/v1/applications/:id/reject

POST   /api/v1/referrals/partners               # Register partner
POST   /api/v1/referrals/links                  # Generate referral link
GET    /api/v1/referrals/leaderboard
GET    /api/v1/referrals/partners/:id/stats
POST   /api/v1/referrals/partners/:id/payout
```
