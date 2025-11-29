# Hướng dẫn Chi tiết - AI Chatbot System

## 📋 Tổng quan

Đây là hệ thống chatbot AI được xây dựng với:
- **Kiến trúc**: Hexagonal Architecture (Clean Architecture)
- **Pattern**: Strategy Pattern cho AI providers
- **Ngôn ngữ**: Golang
- **Database**: MySQL
- **Real-time**: WebSocket
- **AI Models**: Claude (Anthropic), DeepSeek

## 🏗️ Kiến trúc Hexagonal

```
┌─────────────────────────────────────────────────────────────┐
│                      Primary Adapters                       │
│  (Input/Driving - HTTP Handlers, WebSocket Handlers)       │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                        Core Domain                          │
│  ┌──────────────┐    ┌──────────────┐   ┌───────────────┐ │
│  │   Entities   │    │  Use Cases   │   │  Ports (I/F)  │ │
│  │  (Message,   │◄───│ (ChatService)│◄──│ (Interfaces)  │ │
│  │Conversation) │    │              │   │               │ │
│  └──────────────┘    └──────────────┘   └───────────────┘ │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                    Secondary Adapters                       │
│  (Output/Driven - DB Repos, AI Clients)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │   MySQL      │  │Claude Client │  │DeepSeek Client  │ │
│  │ Repositories │  │ (Strategy)   │  │  (Strategy)     │ │
│  └──────────────┘  └──────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 🎯 Strategy Pattern cho AI Providers

Strategy Pattern cho phép chuyển đổi linh hoạt giữa các AI models:

```go
// 1. Định nghĩa Interface (Strategy)
type AIProvider interface {
    SendMessage(ctx, messages, userMessage) (string, error)
    GetModelName() string
}

// 2. Implement các Concrete Strategies
type ClaudeProvider struct { ... }
type DeepSeekProvider struct { ... }

// 3. Factory quản lý strategies
type AIProviderFactory struct {
    providers map[string]AIProvider
}

// 4. Sử dụng
provider := factory.GetProvider("claude")
response := provider.SendMessage(...)
```

## 📁 Cấu trúc Project

```
chatbot-system/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point, dependency injection
├── internal/
│   ├── core/                       # Domain Layer (Business Logic)
│   │   ├── domain/
│   │   │   └── entities.go         # Entities: Message, Conversation
│   │   ├── ports/
│   │   │   └── ports.go            # Interfaces (Ports)
│   │   └── services/
│   │       ├── chat_service.go     # Use Cases
│   │       └── ai_factory.go       # Strategy Factory
│   ├── adapters/
│   │   ├── primary/                # Input Adapters
│   │   │   ├── http/
│   │   │   │   └── handler.go      # REST API handlers
│   │   │   └── websocket/
│   │   │       └── handler.go      # WebSocket handlers
│   │   └── secondary/              # Output Adapters
│   │       ├── ai/
│   │       │   ├── claude_provider.go    # Claude Strategy
│   │       │   └── deepseek_provider.go  # DeepSeek Strategy
│   │       ├── repository/
│   │       │   ├── message_repository.go
│   │       │   └── conversation_repository.go
│   │       └── database/
│   │           └── mysql.go
│   └── config/
│       └── config.go               # Configuration
├── go.mod
├── .env.example
├── docker-compose.yml
├── Makefile
└── demo-client.html
```

## 🚀 Cài đặt và Chạy

### Bước 1: Clone và cài đặt dependencies
```bash
cd chatbot-system
go mod download
```

### Bước 2: Setup MySQL
```bash
docker-compose up -d
```

Hoặc sử dụng MySQL có sẵn, tạo database:
```sql
CREATE DATABASE chatbot_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### Bước 3: Cấu hình Environment
```bash
cp .env.example .env
```

Chỉnh sửa `.env`:
```env
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=chatbot_db

CLAUDE_API_KEY=sk-ant-xxx
CLAUDE_MODEL=claude-sonnet-4-20250514

DEEPSEEK_API_KEY=sk-xxx
DEEPSEEK_MODEL=deepseek-chat
```

### Bước 4: Chạy ứng dụng
```bash
go run cmd/api/main.go
```

Hoặc sử dụng Makefile:
```bash
make run
```

### Bước 5: Test WebSocket
Mở `demo-client.html` trong browser hoặc test với curl:

**REST API:**
```bash
# Tạo conversation
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "title": "Test Chat",
    "ai_model": "claude"
  }'

# Gửi message
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": 1,
    "user_id": "user_123",
    "message": "Hello AI!",
    "ai_model": "claude"
  }'

# Lấy history
curl http://localhost:8080/api/v1/conversations/1/history
```

## 🔄 Flow hoạt động

### Chat Message Flow:
```
1. User gửi message qua WebSocket
   ↓
2. WebSocket Handler nhận message
   ↓
3. Gọi ChatService.ProcessMessage() (Use Case)
   ↓
4. ChatService:
   - Lưu user message vào DB
   - Lấy conversation history
   - Dùng Factory để lấy AI Provider (Strategy Pattern)
   - Gọi AI Provider.SendMessage()
   ↓
5. AI Provider (Claude/DeepSeek) xử lý và trả response
   ↓
6. ChatService lưu AI response vào DB
   ↓
7. Trả response về WebSocket Handler
   ↓
8. WebSocket gửi response cho User
```

## 🎨 Tính năng chính

### 1. Strategy Pattern cho AI Models
- Dễ dàng thêm AI provider mới
- Chuyển đổi model trong cùng conversation
- Cấu hình flexible

### 2. WebSocket Real-time Chat
- Kết nối persistent
- Nhận response ngay lập tức
- Auto-reconnect

### 3. Lưu trữ History
- Lưu toàn bộ lịch sử chat
- Truy vấn theo conversation
- Hỗ trợ context cho AI

### 4. Clean Architecture
- Tách biệt business logic và infrastructure
- Dễ test và maintain
- Tuân thủ SOLID principles

## 🧪 Thêm AI Provider mới

Ví dụ thêm OpenAI GPT:

```go
// 1. Tạo file: internal/adapters/secondary/ai/openai_provider.go
type OpenAIProvider struct {
    apiKey string
    model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
    return &OpenAIProvider{apiKey: apiKey, model: model}
}

func (o *OpenAIProvider) GetModelName() string {
    return "openai"
}

func (o *OpenAIProvider) SendMessage(ctx context.Context, history []domain.Message, userMessage string) (string, error) {
    // Implement OpenAI API call
    // ...
}

// 2. Đăng ký trong main.go
if cfg.OpenAIAPIKey != "" {
    openaiProvider := ai.NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIModel)
    aiFactory.RegisterProvider("openai", openaiProvider)
}
```

## 📊 Database Schema

```sql
-- Conversations table
CREATE TABLE conversations (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  title VARCHAR(255) NOT NULL,
  ai_model VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id)
);

-- Messages table
CREATE TABLE messages (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  conversation_id BIGINT UNSIGNED NOT NULL,
  role VARCHAR(20) NOT NULL,
  content TEXT NOT NULL,
  ai_model VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_conversation_id (conversation_id),
  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);
```

## 🔒 Production Considerations

1. **Security:**
   - Thêm authentication/authorization
   - Validate và sanitize inputs
   - Rate limiting
   - Cấu hình CORS đúng cách

2. **Performance:**
   - Database indexing
   - Caching (Redis)
   - Connection pooling
   - Load balancing

3. **Monitoring:**
   - Logging (structured logging)
   - Metrics (Prometheus)
   - Tracing (Jaeger)
   - Error tracking (Sentry)

4. **Scaling:**
   - Horizontal scaling với multiple instances
   - WebSocket sticky sessions
   - Message queue cho async processing

## 🐛 Troubleshooting

**Lỗi kết nối MySQL:**
```bash
# Kiểm tra MySQL đang chạy
docker-compose ps

# Xem logs
docker-compose logs mysql
```

**WebSocket không kết nối được:**
- Kiểm tra firewall
- Kiểm tra port 8080 available
- Xem browser console logs

**AI Provider error:**
- Kiểm tra API keys trong .env
- Verify API key còn valid
- Kiểm tra rate limits

## 📚 Tài liệu thêm

- [API Documentation](./API_DOCUMENTATION.md)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Strategy Pattern](https://refactoring.guru/design-patterns/strategy)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
