# API Documentation

## WebSocket API

### Kết nối WebSocket
```
ws://localhost:8080/ws?user_id=<user_id>
```

### Message Types

#### 1. Gửi tin nhắn chat
```json
{
  "type": "message",
  "conversation_id": 0,
  "user_id": "user_123",
  "content": "Hello, how are you?",
  "ai_model": "claude"
}
```

**Response:**
```json
{
  "type": "response",
  "conversation_id": 1,
  "message_id": 2,
  "role": "assistant",
  "content": "I'm doing well, thank you for asking!",
  "ai_model": "claude",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### 2. Lấy lịch sử chat
```json
{
  "type": "history",
  "conversation_id": 1
}
```

**Response:**
```json
{
  "type": "history",
  "messages": [
    {
      "id": 1,
      "conversation_id": 1,
      "role": "user",
      "content": "Hello",
      "ai_model": "claude",
      "created_at": "2024-01-15T10:29:00Z"
    },
    {
      "id": 2,
      "conversation_id": 1,
      "role": "assistant",
      "content": "Hi there!",
      "ai_model": "claude",
      "created_at": "2024-01-15T10:29:05Z"
    }
  ]
}
```

#### 3. Chuyển đổi AI model
```json
{
  "type": "switch_model",
  "conversation_id": 1,
  "ai_model": "deepseek"
}
```

**Response:**
```json
{
  "type": "model_switched",
  "ai_model": "deepseek",
  "message": "AI model switched successfully"
}
```

## REST API

### 1. Tạo conversation mới
**POST** `/api/v1/conversations`

**Request:**
```json
{
  "user_id": "user_123",
  "title": "My First Chat",
  "ai_model": "claude"
}
```

**Response:**
```json
{
  "id": 1,
  "user_id": "user_123",
  "title": "My First Chat",
  "ai_model": "claude",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### 2. Lấy lịch sử chat
**GET** `/api/v1/conversations/:conversation_id/history`

**Response:**
```json
{
  "messages": [
    {
      "id": 1,
      "conversation_id": 1,
      "role": "user",
      "content": "Hello",
      "ai_model": "claude",
      "created_at": "2024-01-15T10:29:00Z"
    }
  ]
}
```

### 3. Gửi tin nhắn (REST alternative)
**POST** `/api/v1/chat`

**Request:**
```json
{
  "conversation_id": 1,
  "user_id": "user_123",
  "message": "What's the weather like?",
  "ai_model": "claude"
}
```

**Response:**
```json
{
  "conversation_id": 1,
  "message_id": 3,
  "role": "assistant",
  "content": "I don't have access to real-time weather data...",
  "ai_model": "claude",
  "timestamp": "2024-01-15T10:35:00Z"
}
```

### 4. Chuyển đổi AI model
**PUT** `/api/v1/conversations/:conversation_id/model`

**Request:**
```json
{
  "ai_model": "deepseek"
}
```

**Response:**
```json
{
  "message": "AI model switched successfully"
}
```

### 5. Health Check
**GET** `/health`

**Response:**
```json
{
  "status": "healthy"
}
```

## Database Schema

### Table: conversations
```sql
CREATE TABLE conversations (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  title VARCHAR(255) NOT NULL,
  ai_model VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id)
);
```

### Table: messages
```sql
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

## Strategy Pattern Implementation

Hệ thống sử dụng Strategy Pattern để cho phép chuyển đổi linh hoạt giữa các AI providers:

```go
// Interface chung cho tất cả AI providers
type AIProvider interface {
    SendMessage(ctx context.Context, messages []domain.Message, userMessage string) (string, error)
    GetModelName() string
}

// Factory quản lý các strategies
type AIProviderFactory struct {
    providers map[string]ports.AIProvider
}

// Đăng ký provider
factory.RegisterProvider("claude", claudeProvider)
factory.RegisterProvider("deepseek", deepSeekProvider)

// Sử dụng provider
provider, err := factory.GetProvider("claude")
response, err := provider.SendMessage(ctx, history, message)
```

## Cách chạy ứng dụng

### 1. Cài đặt dependencies
```bash
go mod download
```

### 2. Khởi động MySQL
```bash
docker-compose up -d
```

### 3. Cấu hình environment
```bash
cp .env.example .env
# Chỉnh sửa .env với API keys của bạn
```

### 4. Chạy ứng dụng
```bash
go run cmd/api/main.go
```

hoặc sử dụng Makefile:
```bash
make run
```

### 5. Mở demo client
Mở file `demo-client.html` trong browser để test WebSocket chat.

## Ví dụ sử dụng với JavaScript

```javascript
// Kết nối WebSocket
const ws = new WebSocket('ws://localhost:8080/ws?user_id=user_123');

// Gửi tin nhắn
ws.send(JSON.stringify({
  type: 'message',
  conversation_id: 0,
  user_id: 'user_123',
  content: 'Hello AI!',
  ai_model: 'claude'
}));

// Nhận phản hồi
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('AI Response:', data.content);
};

// Chuyển đổi model
ws.send(JSON.stringify({
  type: 'switch_model',
  conversation_id: 1,
  ai_model: 'deepseek'
}));
```
