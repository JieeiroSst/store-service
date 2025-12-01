# 🤖 Call Center AI System

Hệ thống call center tự động sử dụng AI để trả lời cuộc gọi, chuyển đổi văn bản sang giọng nói, và lưu trữ lịch sử hội thoại.

## ✨ Tính năng

- ☎️ **Tự động trả lời cuộc gọi**: Tích hợp Twilio để nhận và xử lý cuộc gọi
- 🤖 **AI thông minh**: Sử dụng OpenAI GPT hoặc Anthropic Claude
- 🗣️ **Text-to-Speech**: Hỗ trợ Google TTS, ElevenLabs, và gTTS
- 🎤 **Speech-to-Text**: Nhận diện giọng nói tiếng Việt qua Twilio
- 💾 **Lưu trữ lịch sử**: MySQL database lưu tất cả cuộc gọi và tin nhắn
- 📊 **Thống kê và Analytics**: Báo cáo chi tiết về cuộc gọi
- 🎭 **Kịch bản linh hoạt**: Hỗ trợ nhiều kịch bản hội thoại khác nhau
- 🌐 **REST API đầy đủ**: Quản lý cuộc gọi, kịch bản, khách hàng

## 🏗️ Kiến trúc

```
┌──────────────┐
│   Twilio     │ ← Cuộc gọi từ khách hàng
└──────┬───────┘
       │
┌──────▼───────────────┐
│   FastAPI Server     │
├──────────────────────┤
│ • Voice Routes       │
│ • Speech Processing  │
│ • AI Integration     │
│ • TTS Generation     │
└──────┬───────────────┘
       │
┌──────▼───────┐
│  MySQL DB    │ ← Lưu trữ lịch sử
└──────────────┘
```
## Detail
```
┌─────────────┐
│   Client    │ (Điện thoại)
└──────┬──────┘
       │
┌──────▼──────────┐
│   Twilio API    │ (Voice handling)
└──────┬──────────┘
       │
┌──────▼──────────┐
│  FastAPI Server │
├─────────────────┤
│ - Voice Routes  │
│ - TTS/STT       │
│ - AI Processing │
└──────┬──────────┘
       │
┌──────▼──────────┐
│   MySQL DB      │ (Lịch sử cuộc gọi)
└─────────────────┘
```

## 📋 Yêu cầu

- Python 3.11+
- MySQL 8.0+
- Twilio Account (để nhận cuộc gọi)
- OpenAI API Key hoặc Anthropic API Key
- Google Cloud TTS hoặc ElevenLabs API (tùy chọn)

## 🚀 Cài đặt

### 1. Clone repository

```bash
git clone <repository-url>
cd call_center_ai
```

### 2. Tạo virtual environment

```bash
python -m venv venv
source venv/bin/activate  # Linux/Mac
# hoặc
venv\Scripts\activate  # Windows
```

### 3. Cài đặt dependencies

```bash
pip install -r requirements.txt
```

### 4. Cấu hình MySQL

Tạo database:

```sql
CREATE DATABASE call_center_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'callcenter'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON call_center_db.* TO 'callcenter'@'localhost';
FLUSH PRIVILEGES;
```

### 5. Cấu hình biến môi trường

Sao chép file `.env.example` thành `.env` và điền thông tin:

```bash
cp .env.example .env
```

Chỉnh sửa file `.env`:

```env
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=callcenter
DB_PASSWORD=your_password
DB_NAME=call_center_db

# Twilio
TWILIO_ACCOUNT_SID=your_twilio_sid
TWILIO_AUTH_TOKEN=your_twilio_token
TWILIO_PHONE_NUMBER=+1234567890

# AI (chọn 1 trong 2)
OPENAI_API_KEY=your_openai_key
# hoặc
ANTHROPIC_API_KEY=your_anthropic_key

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8000
BASE_URL=https://your-domain.ngrok.io
```

### 6. Khởi tạo database

```bash
python init_scenarios.py
```

### 7. Chạy server

```bash
python main.py
```

Server sẽ chạy tại `http://localhost:8000`

## 🐳 Cài đặt với Docker

### 1. Sử dụng Docker Compose

```bash
# Tạo file .env từ template
cp .env.example .env

# Chỉnh sửa .env với thông tin của bạn

# Khởi động services
docker-compose up -d

# Xem logs
docker-compose logs -f

# Dừng services
docker-compose down
```

### 2. Khởi tạo scenarios trong Docker

```bash
docker-compose exec api python init_scenarios.py
```

## 🔧 Cấu hình Twilio

### 1. Tạo Twilio Account

- Đăng ký tại [twilio.com](https://www.twilio.com)
- Mua số điện thoại có khả năng Voice
- Lấy Account SID và Auth Token

### 2. Expose server ra internet

Sử dụng ngrok để tạo public URL:

```bash
ngrok http 8000
```

Copy HTTPS URL (ví dụ: `https://abc123.ngrok.io`)

### 3. Cấu hình Webhook trong Twilio

1. Vào [Twilio Console](https://console.twilio.com)
2. Chọn Phone Numbers → Manage → Active numbers
3. Chọn số điện thoại của bạn
4. Trong phần "Voice & Fax", cấu hình:
   - **A CALL COMES IN**: Webhook
   - **URL**: `https://your-domain.ngrok.io/voice/incoming`
   - **HTTP**: POST
   - **STATUS CALLBACK**: `https://your-domain.ngrok.io/voice/status`
5. Lưu cấu hình

## 📚 API Documentation

### Voice Endpoints

#### POST /voice/incoming
Xử lý cuộc gọi đến từ Twilio

#### POST /voice/process-speech
Xử lý speech input từ người dùng

#### POST /voice/status
Callback khi trạng thái cuộc gọi thay đổi

### REST API Endpoints

#### Calls Management

```bash
# Lấy danh sách cuộc gọi
GET /api/calls?skip=0&limit=100&status=completed

# Lấy chi tiết cuộc gọi
GET /api/calls/{call_id}

# Xóa cuộc gọi
DELETE /api/calls/{call_id}
```

#### Scenarios Management

```bash
# Lấy danh sách kịch bản
GET /api/scenarios

# Tạo kịch bản mới
POST /api/scenarios
{
  "name": "custom_scenario",
  "description": "Kịch bản tùy chỉnh",
  "prompt": "System prompt cho AI...",
  "is_active": 1
}

# Lấy chi tiết kịch bản
GET /api/scenarios/{scenario_id}

# Cập nhật kịch bản
PUT /api/scenarios/{scenario_id}

# Xóa kịch bản
DELETE /api/scenarios/{scenario_id}
```

#### Customers Management

```bash
# Lấy thông tin khách hàng
GET /api/customers/{phone_number}
```

#### Analytics

```bash
# Lấy thống kê (30 ngày gần nhất)
GET /api/analytics/statistics?days=30
```

### Test Endpoints

```bash
# Test TTS
POST /api/test/tts?text=Xin chào&language=vi-VN
```

## 🎭 Kịch bản có sẵn

Hệ thống có 4 kịch bản mặc định:

1. **customer_support**: Hỗ trợ khách hàng
2. **sales**: Tư vấn bán hàng
3. **appointment**: Đặt lịch hẹn
4. **survey**: Khảo sát ý kiến

Bạn có thể tạo kịch bản tùy chỉnh qua API.

## 💾 Database Schema

### Table: calls
Lưu thông tin cuộc gọi

| Column | Type | Description |
|--------|------|-------------|
| id | INT | Primary key |
| call_sid | VARCHAR(255) | Twilio Call SID |
| from_number | VARCHAR(20) | Số điện thoại gọi đến |
| to_number | VARCHAR(20) | Số điện thoại nhận |
| status | ENUM | Trạng thái cuộc gọi |
| duration | INT | Thời lượng (giây) |
| start_time | DATETIME | Thời gian bắt đầu |
| end_time | DATETIME | Thời gian kết thúc |
| recording_url | VARCHAR(512) | URL file ghi âm |
| notes | TEXT | Ghi chú |

### Table: messages
Lưu tin nhắn trong cuộc hội thoại

| Column | Type | Description |
|--------|------|-------------|
| id | INT | Primary key |
| call_id | INT | Foreign key → calls |
| role | ENUM | user/assistant/system |
| content | TEXT | Nội dung tin nhắn |
| audio_url | VARCHAR(512) | URL file audio |
| timestamp | DATETIME | Thời gian |
| processing_time | FLOAT | Thời gian xử lý |

### Table: scenarios
Lưu kịch bản hội thoại

| Column | Type | Description |
|--------|------|-------------|
| id | INT | Primary key |
| name | VARCHAR(255) | Tên kịch bản |
| description | TEXT | Mô tả |
| prompt | TEXT | System prompt |
| is_active | INT | 1: active, 0: inactive |

### Table: customers
Lưu thông tin khách hàng

| Column | Type | Description |
|--------|------|-------------|
| id | INT | Primary key |
| phone_number | VARCHAR(20) | Số điện thoại |
| name | VARCHAR(255) | Tên khách hàng |
| email | VARCHAR(255) | Email |
| total_calls | INT | Tổng số cuộc gọi |
| last_call_date | DATETIME | Cuộc gọi cuối |

## 🔊 Text-to-Speech Options

Hệ thống hỗ trợ 3 TTS service:

### 1. Google Cloud TTS (Khuyên dùng cho tiếng Việt)
- Chất lượng tốt nhất
- Hỗ trợ nhiều giọng
- Cần Google Cloud credentials

Cài đặt:
```bash
pip install google-cloud-texttospeech
export GOOGLE_APPLICATION_CREDENTIALS="path/to/credentials.json"
```

### 2. ElevenLabs TTS
- Giọng rất tự nhiên
- Hỗ trợ đa ngôn ngữ
- Cần API key

```env
ELEVENLABS_API_KEY=your_api_key
ELEVENLABS_VOICE_ID=voice_id
```

### 3. gTTS (Fallback)
- Miễn phí
- Không cần cấu hình
- Chất lượng cơ bản

```bash
pip install gtts
```

## 🤖 AI Service Options

### 1. Anthropic Claude (Khuyên dùng)
- Hiểu tiếng Việt tốt
- Response tự nhiên
- Chi phí hợp lý

```env
ANTHROPIC_API_KEY=your_api_key
```

### 2. OpenAI GPT
- Model mạnh mẽ
- Đa năng
- Cần API key

```env
OPENAI_API_KEY=your_api_key
```

## 📊 Monitoring và Logs

### View logs trong Docker

```bash
docker-compose logs -f api
```

### Database monitoring

```bash
docker-compose exec mysql mysql -u callcenter -p call_center_db
```

## 🛠️ Troubleshooting

### Lỗi kết nối MySQL

```bash
# Kiểm tra MySQL đang chạy
docker-compose ps

# Restart MySQL
docker-compose restart mysql
```

### Lỗi Twilio webhook

- Kiểm tra ngrok đang chạy
- Đảm bảo BASE_URL trong .env đúng
- Kiểm tra Twilio webhook configuration

### Lỗi TTS

- Kiểm tra API keys
- Thử fallback sang gTTS
- Xem logs để debug

### Lỗi AI response

- Kiểm tra API keys
- Kiểm tra rate limits
- Xem conversation history

## 🔐 Security Best Practices

1. Không commit file `.env`
2. Sử dụng strong passwords cho database
3. Giới hạn access đến API endpoints
4. Sử dụng HTTPS cho production
5. Rotate API keys định kỳ
6. Implement rate limiting

## 📈 Performance Optimization

1. Sử dụng connection pooling cho database
2. Cache AI responses cho câu hỏi phổ biến
3. Compress audio files
4. Implement CDN cho static files
5. Scale horizontally với load balancer

## 🤝 Contributing

Mọi đóng góp đều được hoan nghênh! Vui lòng:

1. Fork repository
2. Tạo feature branch
3. Commit changes
4. Push to branch
5. Tạo Pull Request

## 📝 License

MIT License

## 📧 Contact

Để được hỗ trợ, vui lòng tạo issue trên GitHub.

## 🎯 Roadmap

- [ ] Thêm multi-language support
- [ ] Implement queue system
- [ ] Add web dashboard
- [ ] Real-time analytics
- [ ] Call recording và transcription
- [ ] Integration với CRM systems
- [ ] Advanced AI training với custom data
- [ ] Mobile app

---

**Made with ❤️ for Vietnamese Call Centers**
