# AI Chatbot System - Hexagonal Architecture

## Tổng quan

Hệ thống chatbot với AI sử dụng **Strategy Pattern** để chuyển đổi giữa các AI models (Claude, DeepSeek, etc.) và **Hexagonal Clean Architecture** bằng Golang.

### Tính năng chính

- ✅ **WebSocket real-time chat** - Chat thời gian thực
- ✅ **Multi AI provider** - Hỗ trợ nhiều AI (Claude, DeepSeek) với Strategy Pattern
- ✅ **MySQL database** - Lưu trữ lịch sử chat
- ✅ **Phân quyền phức tạp**:
  - Manager: có thể chat với n users (1-n)
  - Advisor: có thể chat với 1 user (1-1)
  - User: có thể chat với manager hoặc advisor của mình
- ✅ **Clean Architecture** - Hexagonal Architecture pattern

## Kiến trúc

```
┌─────────────────────────────────────────────────────────────┐
│                     Infrastructure Layer                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │   HTTP   │  │WebSocket │  │ Database │  │    AI    │   │
│  │ Handlers │  │ Handlers │  │  (MySQL) │  │Providers │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│                    ┌──────────────┐                          │
│                    │  Use Cases   │                          │
│                    │ (ChatUseCase)│                          │
│                    └──────────────┘                          │
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │
┌─────────────────────────────────────────────────────────────┐
│                       Domain Layer                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Entities │  │Interfaces│  │  Value   │  │ Business │   │
│  │(User,Msg)│  │(Repos)   │  │ Objects  │  │  Rules   │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Strategy Pattern - AI Providers

```go
// Domain Interface
type AIProvider interface {
    SendMessage(ctx, conversation, userMessage) (string, error)
    GetModelName() string
    IsAvailable() bool
}

// Concrete Strategies
- ClaudeProvider
- DeepSeekProvider
- (Dễ dàng thêm providers mới)
```

## Cấu trúc thư mục

```
chatbot-system/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point
├── internal/
│   ├── domain/                     # Domain Layer
│   │   ├── user.go                # User entity
│   │   ├── message.go             # Message, Conversation entities
│   │   ├── ai_provider.go         # AI Strategy interface
│   │   └── repository.go          # Repository interfaces
│   ├── application/                # Application Layer
│   │   └── chat_usecase.go        # Business logic
│   ├── infrastructure/             # Infrastructure Layer
│   │   ├── ai/                    # AI Adapters
│   │   │   ├── claude_provider.go
│   │   │   ├── deepseek_provider.go
│   │   │   └── factory.go
│   │   ├── database/              # Database Adapters
│   │   │   ├── user_repository.go
│   │   │   └── message_repository.go
│   │   ├── websocket/             # WebSocket Adapters
│   │   │   ├── hub.go
│   │   │   └── handler.go
│   │   └── http/                  # HTTP Adapters
│   │       └── chat_handler.go
│   └── config/
│       └── config.go              # Configuration
├── migrations/
│   └── 001_init_schema.sql       # Database schema
├── go.mod
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Cài đặt và Chạy

### 1. Prerequisites

```bash
- Go 1.21+
- Docker & Docker Compose
- MySQL 8.0+
```

### 2. Clone và Setup

```bash
git clone <repository>
cd chatbot-system

# Copy environment file
cp .env.example .env

# Chỉnh sửa .env với API keys của bạn
nano .env
```

### 3. Chạy với Docker

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f app

# Stop services
docker-compose down
```

### 4. Chạy local (không dùng Docker)

```bash
# Start MySQL
docker-compose up -d mysql

# Install dependencies
go mod download

# Run migrations
mysql -h localhost -u root -p chatbot_db < migrations/001_init_schema.sql

# Run application
export CLAUDE_API_KEY=your_key
export DEEPSEEK_API_KEY=your_key
go run cmd/api/main.go
```

## API Documentation

### REST APIs

#### 1. Get Conversation History
```http
GET /api/conversations/{conversation_id}/history?user_id=1&limit=50&offset=0
```

**Response:**
```json
{
  "messages": [
    {
      "id": 1,
      "conversation_id": 1,
      "sender_id": 1,
      "content": "Hello!",
      "message_type": "user",
      "created_at": "2024-01-01T10:00:00Z"
    }
  ],
  "count": 1
}
```

#### 2. Get User Conversations
```http
GET /api/conversations?user_id=1
```

**Response:**
```json
{
  "conversations": [
    {
      "id": 1,
      "user1_id": 1,
      "user2_id": 0,
      "is_ai_chat": true,
      "active_ai_model": "claude",
      "created_at": "2024-01-01T10:00:00Z"
    }
  ],
  "count": 1
}
```

### WebSocket API

#### Kết nối WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?user_id=1');
```

#### 1. Send Message (Human to Human)
```json
{
  "type": "send_message",
  "payload": {
    "recipient_id": 2,
    "content": "Hello from user 1",
    "conversation_id": 1
  }
}
```

#### 2. Send Message (AI Chat)
```json
{
  "type": "send_message",
  "payload": {
    "content": "Hello AI",
    "ai_model": "claude"
  }
}
```

#### 3. Switch AI Model
```json
{
  "type": "switch_ai_model",
  "payload": {
    "conversation_id": 1,
    "new_model": "deepseek"
  }
}
```

#### WebSocket Response
```json
{
  "type": "message",
  "message_id": 1,
  "sender_id": 0,
  "content": "AI response here",
  "message_type": "ai",
  "ai_model": "claude",
  "conversation_id": 1
}
```

## Database Schema

### Users Table
```sql
- id (BIGINT, PK)
- username (VARCHAR)
- email (VARCHAR, UNIQUE)
- role (ENUM: manager, advisor, user)
- manager_id (BIGINT, FK -> users.id)
- advisor_id (BIGINT, FK -> users.id)
- created_at, updated_at
```

### Conversations Table
```sql
- id (BIGINT, PK)
- user1_id (BIGINT, FK)
- user2_id (BIGINT, FK)
- is_ai_chat (BOOLEAN)
- active_ai_model (VARCHAR)
- created_at, updated_at
```

### Messages Table
```sql
- id (BIGINT, PK)
- conversation_id (BIGINT, FK)
- sender_id (BIGINT)
- content (TEXT)
- message_type (ENUM: user, ai)
- ai_model (VARCHAR)
- created_at
```

## Mô hình phân quyền

### Sample Data
```
Manager1 (id=1)
  └─ manages -> User1 (id=3), User2 (id=4)

Advisor1 (id=2)
  └─ advises -> User1 (id=3)

User1 (id=3)
  ├─ can chat with Manager1
  └─ can chat with Advisor1

User2 (id=4)
  └─ can chat with Manager1
```

### Business Rules
- Manager có thể chat với nhiều users mà họ quản lý
- Advisor chỉ có thể chat với 1 user được assign
- User có thể chat với manager hoặc advisor của mình
- User có thể chat với AI bất cứ lúc nào

## Testing

### Test WebSocket với JavaScript

```html
<!DOCTYPE html>
<html>
<body>
<script>
const ws = new WebSocket('ws://localhost:8080/ws?user_id=1');

ws.onopen = () => {
  console.log('Connected');
  
  // Send message to AI
  ws.send(JSON.stringify({
    type: 'send_message',
    payload: {
      content: 'Hello Claude!',
      ai_model: 'claude'
    }
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
</script>
</body>
</html>
```

## Thêm AI Provider mới

### 1. Tạo provider implementation
```go
// internal/infrastructure/ai/openai_provider.go
type OpenAIProvider struct {
    apiKey string
}

func (o *OpenAIProvider) SendMessage(ctx context.Context, 
    conversation []domain.Message, userMessage string) (string, error) {
    // Implementation
}

func (o *OpenAIProvider) GetModelName() string {
    return "openai"
}

func (o *OpenAIProvider) IsAvailable() bool {
    return o.apiKey != ""
}
```

### 2. Thêm vào Factory
```go
// internal/infrastructure/ai/factory.go
func NewProviderFactory(claudeKey, deepseekKey, openaiKey string) *ProviderFactory {
    // ...
    if openaiKey != "" {
        factory.providers["openai"] = NewOpenAIProvider(openaiKey)
    }
}
```

## Troubleshooting

### Database connection error
```bash
# Check MySQL is running
docker-compose ps

# View MySQL logs
docker-compose logs mysql

# Connect to MySQL
docker exec -it chatbot_mysql mysql -u root -p
```

### WebSocket connection failed
```bash
# Check if server is running
curl http://localhost:8080/health

# Check logs
docker-compose logs -f app
```

## Production Considerations

1. **Authentication**: Implement JWT or session-based auth
2. **Rate Limiting**: Add rate limiting cho API và WebSocket
3. **Logging**: Implement structured logging
4. **Monitoring**: Add metrics và health checks
5. **CORS**: Configure CORS properly cho production
6. **Database**: Thêm connection pooling, indices optimization
7. **Security**: Hash passwords, validate inputs, sanitize data

## License

MIT

## Contact

Nếu có câu hỏi, vui lòng tạo issue hoặc liên hệ.
