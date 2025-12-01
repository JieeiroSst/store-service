# 🚀 Quick Start Guide

Hướng dẫn setup nhanh hệ thống Call Center AI trong 10 phút!

## 📋 Checklist trước khi bắt đầu

- [ ] Python 3.11+ đã cài đặt
- [ ] MySQL 8.0+ đã cài đặt và đang chạy
- [ ] Đã có Twilio account (hoặc đang trial)
- [ ] Đã có OpenAI hoặc Anthropic API key
- [ ] (Tùy chọn) Docker và Docker Compose

## 🎯 Option 1: Setup thủ công (Local)

### Bước 1: Clone và cài đặt

```bash
# Clone repository
cd call_center_ai

# Tạo virtual environment
python -m venv venv
source venv/bin/activate  # Linux/Mac
# hoặc: venv\Scripts\activate  # Windows

# Cài đặt dependencies
pip install -r requirements.txt
```

### Bước 2: Setup MySQL

```sql
-- Chạy trong MySQL client
CREATE DATABASE call_center_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'callcenter'@'localhost' IDENTIFIED BY 'CallCenter2024!';
GRANT ALL PRIVILEGES ON call_center_db.* TO 'callcenter'@'localhost';
FLUSH PRIVILEGES;
```

### Bước 3: Cấu hình môi trường

```bash
# Copy file mẫu
cp .env.example .env

# Chỉnh sửa .env (dùng nano, vim, hoặc editor bất kỳ)
nano .env
```

**Minimum configuration trong .env:**

```env
# Database (bắt buộc)
DB_HOST=localhost
DB_PORT=3306
DB_USER=callcenter
DB_PASSWORD=CallCenter2024!
DB_NAME=call_center_db

# Twilio (bắt buộc cho call)
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_PHONE_NUMBER=+1234567890

# AI - Chọn 1 trong 2 (bắt buộc)
ANTHROPIC_API_KEY=sk-ant-xxxxx
# hoặc
# OPENAI_API_KEY=sk-xxxxx

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8000
```

### Bước 4: Khởi tạo database và scenarios

```bash
python init_scenarios.py
```

### Bước 5: Test hệ thống

```bash
# Test các thành phần
python test_system.py

# Nếu muốn test cả API endpoints
python test_system.py --with-api
```

### Bước 6: Chạy server

```bash
python main.py
```

Server chạy tại: `http://localhost:8000`

### Bước 7: Expose server ra internet

```bash
# Terminal mới, cài ngrok nếu chưa có
# Download từ: https://ngrok.com/download

# Chạy ngrok
ngrok http 8000
```

Copy HTTPS URL (ví dụ: `https://abc123.ngrok.io`)

### Bước 8: Cấu hình Twilio

1. Vào [Twilio Console](https://console.twilio.com/us1/develop/phone-numbers/manage/incoming)
2. Chọn số điện thoại của bạn
3. Trong phần "Voice Configuration":
   - **A CALL COMES IN**: Webhook, POST
   - **URL**: `https://abc123.ngrok.io/voice/incoming`
   - **STATUS CALLBACK URL**: `https://abc123.ngrok.io/voice/status`
4. Save

### Bước 9: Test cuộc gọi! 🎉

Gọi đến số Twilio của bạn và nói chuyện với AI!

---

## 🐳 Option 2: Setup với Docker (Khuyên dùng)

### Bước 1: Chuẩn bị

```bash
# Đảm bảo Docker và Docker Compose đã cài
docker --version
docker-compose --version

# Copy .env
cp .env.example .env
```

### Bước 2: Cấu hình .env

Chỉnh sửa `.env` với thông tin của bạn (giống Option 1)

### Bước 3: Khởi động

```bash
# Build và start tất cả services
docker-compose up -d

# Xem logs
docker-compose logs -f

# Đợi MySQL khởi động (khoảng 30 giây)
```

### Bước 4: Khởi tạo scenarios

```bash
docker-compose exec api python init_scenarios.py
```

### Bước 5: Test

```bash
# Test trong container
docker-compose exec api python test_system.py
```

### Bước 6: Expose với ngrok

```bash
ngrok http 8000
```

### Bước 7: Cấu hình Twilio

Giống như Option 1, bước 8

### Bước 8: Test cuộc gọi! 🎉

---

## 🔍 Kiểm tra hệ thống

### Health Check

```bash
curl http://localhost:8000/
```

Kết quả mong đợi:
```json
{
  "status": "healthy",
  "service": "Call Center AI",
  "version": "1.0.0"
}
```

### Test TTS

```bash
curl -X POST "http://localhost:8000/api/test/tts?text=Xin chào&language=vi-VN" \
  --output test.mp3
```

### Xem danh sách cuộc gọi

```bash
curl http://localhost:8000/api/calls
```

### Xem scenarios

```bash
curl http://localhost:8000/api/scenarios
```

### Xem thống kê

```bash
curl http://localhost:8000/api/analytics/statistics?days=30
```

---

## 🆘 Troubleshooting nhanh

### ❌ Lỗi: "Can't connect to MySQL"

```bash
# Kiểm tra MySQL đang chạy
sudo systemctl status mysql  # Linux
# hoặc
brew services list  # Mac

# Start MySQL nếu cần
sudo systemctl start mysql  # Linux
brew services start mysql  # Mac
```

### ❌ Lỗi: "No AI service available"

- Kiểm tra API key trong .env
- Đảm bảo có ít nhất 1 AI service (OpenAI hoặc Anthropic)
- Test API key:

```bash
# Test OpenAI
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Test Anthropic
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01"
```

### ❌ Lỗi: "No TTS service available"

Hệ thống sẽ tự động fallback sang gTTS nếu không có TTS service nào khác:

```bash
pip install gtts
```

### ❌ Twilio không gọi được webhook

1. Kiểm tra ngrok đang chạy:
   ```bash
   curl https://your-url.ngrok.io/
   ```

2. Kiểm tra Twilio webhook URL trong console

3. Xem logs:
   ```bash
   # Local
   tail -f nohup.out
   
   # Docker
   docker-compose logs -f api
   ```

### ❌ Port 8000 đã được sử dụng

```bash
# Tìm process đang dùng port 8000
lsof -i :8000  # Linux/Mac
netstat -ano | findstr :8000  # Windows

# Kill process hoặc đổi port trong .env
SERVER_PORT=8001
```

---

## 📱 Test Call Flow

### Kịch bản test mẫu:

1. **Gọi đến số Twilio**
2. **Bot chào**: "Xin chào! Tôi là trợ lý ảo của công ty..."
3. **Bạn nói**: "Tôi cần hỗ trợ về sản phẩm"
4. **Bot trả lời**: AI sẽ phản hồi phù hợp
5. **Tiếp tục hội thoại** hoặc nói "tạm biệt" để kết thúc

### Kiểm tra lịch sử:

```bash
# Xem cuộc gọi mới nhất
curl http://localhost:8000/api/calls?limit=1

# Xem chi tiết cuộc gọi (thay {call_id})
curl http://localhost:8000/api/calls/1
```

---

## 🎓 Các bước tiếp theo

✅ Hệ thống đã chạy? Tuyệt vời! Bây giờ bạn có thể:

1. **Tùy chỉnh kịch bản**: Sửa scenarios trong database
2. **Thêm TTS tốt hơn**: Cấu hình Google TTS hoặc ElevenLabs
3. **Deploy lên production**: Sử dụng server thật thay vì ngrok
4. **Tích hợp CRM**: Kết nối với hệ thống hiện có
5. **Custom AI training**: Fine-tune responses cho domain cụ thể

---

## 📚 Tài liệu tham khảo

- [README.md](README.md) - Documentation đầy đủ
- [Twilio Voice Docs](https://www.twilio.com/docs/voice)
- [FastAPI Docs](https://fastapi.tiangolo.com/)
- [SQLAlchemy Docs](https://docs.sqlalchemy.org/)

---

## 💡 Tips

1. **Development**: Dùng ngrok free plan
2. **Production**: Deploy lên cloud với domain thật
3. **Security**: Đừng commit file .env
4. **Monitoring**: Setup logging và alerts
5. **Backup**: Backup database thường xuyên

---

**🎉 Chúc mừng! Bạn đã có hệ thống Call Center AI hoàn chỉnh!**

Có câu hỏi? Tạo issue trên GitHub hoặc liên hệ support.
