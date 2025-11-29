# Quick Start - AI Chatbot System

## 🚀 Bắt đầu nhanh (5 phút)

### 1. Cài đặt dependencies
```bash
cd chatbot-system
go mod download
```

### 2. Chạy MySQL
```bash
docker-compose up -d
```

### 3. Cấu hình
```bash
cp .env.example .env
# Chỉnh sửa .env và thêm API keys:
# CLAUDE_API_KEY=sk-ant-xxx
# DEEPSEEK_API_KEY=sk-xxx
```

### 4. Chạy server
```bash
go run cmd/api/main.go
```

### 5. Test
Mở `demo-client.html` trong browser hoặc:

```bash
# Test REST API
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test","title":"Test","ai_model":"claude"}'
```

## 📂 Files quan trọng

- `cmd/api/main.go` - Entry point
- `demo-client.html` - WebSocket demo client
- `SETUP_GUIDE.md` - Hướng dẫn chi tiết
- `API_DOCUMENTATION.md` - API docs

## 🏗️ Kiến trúc

```
Core Domain (Business Logic)
    ├── Entities (Message, Conversation)
    ├── Use Cases (ChatService)
    └── Ports (Interfaces)
         ↕️
Adapters
    ├── Primary (HTTP, WebSocket)
    └── Secondary (MySQL, AI Providers)
```

## 🎯 Strategy Pattern

```go
// Dễ dàng switch giữa AI models
provider := factory.GetProvider("claude")  // hoặc "deepseek"
response := provider.SendMessage(...)
```

## 📝 Features

✅ WebSocket real-time chat
✅ REST API
✅ MySQL lưu history
✅ Strategy Pattern cho AI models
✅ Clean Architecture
✅ Dễ dàng thêm AI provider mới

## 🔧 Thêm AI Provider mới

1. Tạo file mới trong `internal/adapters/secondary/ai/`
2. Implement interface `AIProvider`
3. Đăng ký trong `main.go`

```go
aiFactory.RegisterProvider("new_ai", newProvider)
```

## 🌐 Endpoints

### WebSocket
```
ws://localhost:8080/ws?user_id=<user_id>
```

### REST API
```
POST   /api/v1/conversations              # Tạo conversation
GET    /api/v1/conversations/:id/history  # Lấy history
POST   /api/v1/chat                        # Gửi message
PUT    /api/v1/conversations/:id/model    # Switch model
GET    /health                             # Health check
```

## 📚 Docs đầy đủ

- [SETUP_GUIDE.md](./SETUP_GUIDE.md) - Hướng dẫn chi tiết
- [API_DOCUMENTATION.md](./API_DOCUMENTATION.md) - API reference

## 🎨 Demo Client Usage

1. Mở `demo-client.html` trong browser
2. Chọn AI model (Claude/DeepSeek)
3. Nhập message và chat
4. Có thể switch model trong conversation

## 🛠️ Makefile Commands

```bash
make run          # Chạy ứng dụng
make build        # Build binary
make docker-up    # Start MySQL
make docker-down  # Stop MySQL
```

## ⚡ Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **WebSocket**: gorilla/websocket
- **Database**: MySQL + GORM
- **AI**: Claude API, DeepSeek API
- **Pattern**: Strategy, Hexagonal Architecture
