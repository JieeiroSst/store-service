# 📁 Project Structure

## Tổng quan cấu trúc

```
call_center_ai/
├── 📄 main.py                  # FastAPI application chính
├── 📄 database.py              # Database connection & session management
├── 📄 models.py                # SQLAlchemy models (tables)
├── 📄 schemas.py               # Pydantic schemas (API validation)
├── 📄 crud.py                  # Database CRUD operations
├── 📄 ai_service.py            # AI integration (OpenAI, Anthropic)
├── 📄 tts_service.py           # Text-to-Speech services
├── 📄 init_scenarios.py        # Script khởi tạo scenarios mặc định
├── 📄 test_system.py           # Testing script
│
├── 📄 requirements.txt         # Python dependencies
├── 📄 .env.example             # Environment variables template
├── 📄 .env                     # Environment variables (không commit)
├── 📄 .gitignore              # Git ignore rules
│
├── 🐳 Dockerfile               # Docker image configuration
├── 🐳 docker-compose.yml       # Docker Compose setup
│
├── 📖 README.md                # Documentation đầy đủ
├── 📖 QUICKSTART.md            # Hướng dẫn setup nhanh
└── 📖 PROJECT_STRUCTURE.md     # File này
```

---

## 📄 Chi tiết từng file

### Core Application Files

#### `main.py` (1,200+ lines)
**Mục đích**: FastAPI application chính, xử lý tất cả HTTP routes

**Chức năng chính**:
- Voice endpoints cho Twilio webhooks
- REST API endpoints (calls, scenarios, customers, analytics)
- Lifecycle management (startup/shutdown)
- Middleware configuration (CORS)
- Conversation history management

**Key endpoints**:
```python
POST /voice/incoming          # Xử lý cuộc gọi đến
POST /voice/process-speech    # Xử lý speech input
POST /voice/status           # Callback trạng thái cuộc gọi

GET  /api/calls              # Lấy danh sách cuộc gọi
GET  /api/calls/{id}         # Chi tiết cuộc gọi
GET  /api/scenarios          # Danh sách kịch bản
POST /api/scenarios          # Tạo kịch bản mới
GET  /api/analytics/statistics  # Thống kê
```

---

#### `database.py` (~50 lines)
**Mục đích**: Quản lý kết nối database và session

**Chức năng**:
- SQLAlchemy engine configuration
- Connection pooling setup
- Session factory
- Database initialization
- Dependency injection cho FastAPI

**Configuration**:
```python
- Pool size: 10
- Max overflow: 20
- Pool recycle: 3600s
- Pre-ping: True (kiểm tra connection trước khi dùng)
```

---

#### `models.py` (~150 lines)
**Mục đích**: Định nghĩa database schema bằng SQLAlchemy ORM

**Models**:

1. **Call** - Thông tin cuộc gọi
   - id, call_sid, from/to_number
   - status, duration, start/end_time
   - recording_url, notes
   - Relationship: One-to-Many với Message

2. **Message** - Tin nhắn trong cuộc hội thoại
   - id, call_id, role (user/assistant/system)
   - content, audio_url
   - timestamp, processing_time
   - Relationship: Many-to-One với Call

3. **Scenario** - Kịch bản hội thoại
   - id, name, description
   - prompt (system prompt cho AI)
   - is_active

4. **Customer** - Thông tin khách hàng
   - id, phone_number, name, email
   - total_calls, last_call_date
   - notes

**Enums**:
- CallStatus: initiated, in_progress, completed, failed, no_answer, busy
- MessageRole: user, assistant, system

---

#### `schemas.py` (~120 lines)
**Mục đích**: Pydantic models cho API validation và serialization

**Schema groups**:
- Call schemas: CallBase, CallCreate, CallUpdate, CallResponse
- Message schemas: MessageBase, MessageCreate, MessageResponse
- Scenario schemas: ScenarioBase, ScenarioCreate, ScenarioUpdate, ScenarioResponse
- Customer schemas: CustomerBase, CustomerCreate, CustomerUpdate, CustomerResponse
- Special: CallHistoryResponse (call + messages), CallStatistics

---

#### `crud.py` (~300 lines)
**Mục đích**: Database operations (Create, Read, Update, Delete)

**Operations theo model**:

**Call operations**:
- create_call, get_call, get_call_by_sid
- get_calls (with filters), update_call, delete_call
- get_call_with_messages

**Message operations**:
- create_message, get_messages_by_call

**Scenario operations**:
- create_scenario, get_scenario, get_scenario_by_name
- get_scenarios, update_scenario, delete_scenario

**Customer operations**:
- create_customer, get_customer_by_phone
- get_or_create_customer, update_customer
- increment_customer_calls

**Analytics**:
- get_call_statistics(days) - thống kê cuộc gọi

---

#### `ai_service.py` (~350 lines)
**Mục đích**: Tích hợp AI services để tạo responses

**Services**:

1. **OpenAIService**
   - Sử dụng GPT-4 Turbo
   - Temperature: 0.7
   - Max tokens: 500

2. **AnthropicService**
   - Sử dụng Claude 3 Sonnet
   - Temperature: 0.7
   - Max tokens: 500

**AIManager**:
- Quản lý nhiều AI services
- Fallback mechanism (thử service theo thứ tự)
- Trả về response + processing time

**ScenarioManager**:
- Quản lý 4 kịch bản mặc định:
  1. customer_support - Hỗ trợ khách hàng
  2. sales - Tư vấn bán hàng
  3. appointment - Đặt lịch hẹn
  4. survey - Khảo sát ý kiến

---

#### `tts_service.py` (~300 lines)
**Mục đích**: Text-to-Speech conversion

**Services**:

1. **GoogleTTS** (Priority 1)
   - Google Cloud Text-to-Speech
   - Chất lượng cao nhất
   - Hỗ trợ vi-VN-Standard-A voice

2. **ElevenLabsTTS** (Priority 2)
   - ElevenLabs API
   - Giọng tự nhiên
   - Multilingual v2 model

3. **SimpleTTS** (Fallback)
   - gTTS (Google Translate TTS)
   - Miễn phí, không cần API key
   - Chất lượng cơ bản

**TTSManager**:
- Quản lý nhiều TTS services
- Automatic fallback
- Audio file saving

---

### Utility Files

#### `init_scenarios.py` (~50 lines)
**Mục đích**: Khởi tạo kịch bản mặc định vào database

**Usage**:
```bash
python init_scenarios.py
```

---

#### `test_system.py` (~200 lines)
**Mục đích**: Test tất cả thành phần của hệ thống

**Tests**:
- Database connection
- AI service (generate response)
- TTS service (synthesize audio)
- Scenario management
- API endpoints (optional)

**Usage**:
```bash
python test_system.py
python test_system.py --with-api  # Include API tests
```

---

### Configuration Files

#### `requirements.txt`
**Dependencies chính**:
```
fastapi==0.104.1          # Web framework
uvicorn==0.24.0           # ASGI server
sqlalchemy==2.0.23        # ORM
pymysql==1.1.0            # MySQL driver
twilio==8.10.0            # Twilio SDK
openai==1.3.5             # OpenAI API
anthropic==0.7.7          # Anthropic API
elevenlabs==0.2.27        # ElevenLabs TTS
google-cloud-texttospeech # Google TTS
```

#### `.env.example`
Template cho environment variables:
- Database config
- Twilio credentials
- AI API keys
- Server settings
- TTS configuration

#### `.gitignore`
Ignore rules cho:
- Python artifacts (__pycache__, *.pyc)
- Virtual environment
- .env files
- IDE files
- Logs và temporary files

---

### Docker Files

#### `Dockerfile`
**Base image**: python:3.11-slim

**Layers**:
1. Install system dependencies (gcc, mysql-dev)
2. Copy requirements.txt
3. Install Python packages
4. Copy application code
5. Create audio directory
6. Expose port 8000

---

#### `docker-compose.yml`
**Services**:

1. **mysql**
   - MySQL 8.0
   - Port 3306
   - Volume for data persistence
   - Health check

2. **api**
   - FastAPI application
   - Port 8000
   - Auto-reload enabled
   - Depends on MySQL
   - Volume mounts for code

**Networks**: call_center_network (bridge)

---

### Documentation Files

#### `README.md` (2,500+ lines)
**Comprehensive documentation**:
- Features overview
- Architecture diagram
- Installation (manual & Docker)
- Twilio configuration
- API documentation
- Database schema
- TTS options
- AI service options
- Troubleshooting
- Security best practices
- Performance optimization

---

#### `QUICKSTART.md` (1,000+ lines)
**Quick setup guide**:
- Prerequisites checklist
- Option 1: Manual setup (9 steps)
- Option 2: Docker setup (8 steps)
- System checks
- Quick troubleshooting
- Test call flow
- Next steps

---

## 🔄 Data Flow

### Incoming Call Flow

```
1. Twilio receives call
   ↓
2. POST /voice/incoming
   ↓
3. Create Call record in DB
   ↓
4. Create Customer record (if new)
   ↓
5. Generate greeting with AI
   ↓
6. Convert to speech (TTS)
   ↓
7. Twilio plays audio
   ↓
8. User speaks → Twilio STT
   ↓
9. POST /voice/process-speech
   ↓
10. Save user message to DB
    ↓
11. Get AI response
    ↓
12. Save AI message to DB
    ↓
13. Convert to speech (TTS)
    ↓
14. Return TwiML response
    ↓
15. Loop steps 8-14 or end call
    ↓
16. POST /voice/status (callback)
    ↓
17. Update Call status in DB
```

---

## 🗄️ Database Relationships

```
customers (1) ──< calls (N)
                    │
                    │ (1)
                    │
                    ↓
                  (N) messages

scenarios (standalone)
```

---

## 🔌 External Integrations

### Required
- **Twilio**: Voice calling, Speech-to-Text
- **OpenAI** hoặc **Anthropic**: AI responses

### Optional
- **Google Cloud TTS**: Text-to-Speech
- **ElevenLabs**: Premium TTS
- **gTTS**: Free TTS fallback

---

## 📊 File Metrics

| File | Lines | Purpose |
|------|-------|---------|
| main.py | ~1,200 | FastAPI app & routes |
| tts_service.py | ~300 | TTS integration |
| ai_service.py | ~350 | AI integration |
| crud.py | ~300 | Database operations |
| models.py | ~150 | Database models |
| schemas.py | ~120 | API schemas |
| database.py | ~50 | DB connection |
| test_system.py | ~200 | System testing |
| README.md | ~2,500 | Documentation |
| QUICKSTART.md | ~1,000 | Setup guide |

**Total**: ~6,000+ lines of code & documentation

---

## 🎯 Design Patterns Used

1. **Repository Pattern**: CRUD operations separated in crud.py
2. **Service Pattern**: AI and TTS services encapsulated
3. **Factory Pattern**: TTSManager, AIManager create service instances
4. **Strategy Pattern**: Multiple TTS/AI implementations with common interface
5. **Dependency Injection**: FastAPI's Depends() for database sessions
6. **Singleton Pattern**: Global manager instances (tts_manager, ai_manager)

---

## 🔒 Security Considerations

1. **Environment Variables**: Sensitive data in .env (not committed)
2. **SQL Injection Protection**: SQLAlchemy ORM prevents SQL injection
3. **Input Validation**: Pydantic schemas validate all inputs
4. **Connection Pooling**: Limited connections to prevent resource exhaustion
5. **CORS**: Configurable cross-origin rules

---

## 🚀 Scalability Features

1. **Connection Pooling**: Handles concurrent database connections
2. **Async Processing**: FastAPI async endpoints
3. **Stateless Design**: Can run multiple instances
4. **Fallback Services**: Automatic failover for AI and TTS
5. **Docker Support**: Easy horizontal scaling

---

**Cấu trúc này được thiết kế để**:
- ✅ Dễ maintain và extend
- ✅ Clear separation of concerns
- ✅ Testable và debuggable
- ✅ Production-ready
- ✅ Well-documented
