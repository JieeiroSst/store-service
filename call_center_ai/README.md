# 🤖 Call Center AI System

Hệ thống call center tự động sử dụng AI để trả lời cuộc gọi, chuyển đổi văn bản sang giọng nói, và lưu trữ lịch sử hội thoại.

## ✨ Tính năng

- ☎️ **Tự động trả lời cuộc gọi** qua Twilio (nhận diện giọng nói tiếng Việt, đọc câu trả lời bằng TTS của Twilio)
- 🎭 **6 kịch bản dựng sẵn**: chăm sóc khách hàng, đơn hàng & đổi trả, tư vấn bán hàng, khiếu nại, đặt lịch hẹn, khảo sát; mỗi kịch bản có lời chào và prompt riêng
- 📞 **Định tuyến theo số được gọi**: mỗi số tổng đài (`phone_number`) gắn với một kịch bản, có kịch bản mặc định
- 📚 **Kho kiến thức FAQ**: bot chỉ trả lời dựa trên FAQ khớp với câu hỏi (tìm kiếm tiếng Việt có dấu hoặc không dấu), không bịa chính sách hay giá
- 🙋 **Chuyển máy cho nhân viên**: khi khách yêu cầu, khi AI không trả lời được hoặc khi hội thoại quá dài, bot `<Dial>` tới tổng đài viên; nếu không có ai trực thì bot ghi nhận yêu cầu gọi lại
- 🧠 **Cá nhân hoá**: dùng tên, ghi chú và số lần gọi của khách trong câu trả lời
- 🤖 **AI**: Claude (`claude-opus-5-5`, có server-side refusal fallback) làm provider chính, OpenAI làm dự phòng
- 📊 **Thống kê theo kết quả**: resolved / transferred / callback_requested / abandoned
- 🧪 **Giả lập cuộc gọi** qua `POST /api/simulate` để kiểm thử kịch bản mà không cần gọi điện

## 🏗️ Kiến trúc (Hexagonal / Ports & Adapters)

```
 inbound adapters            application         domain            outbound adapters
 ─────────────────           ───────────         ──────            ─────────────────
 FastAPI /voice (Twilio) ─┐                                  ┌─ SQLAlchemy + MySQL
 FastAPI /api   (admin)  ─┼──▶ use cases ──▶ entities + ports ◀─┼─ Anthropic / OpenAI
 CLI seed_scenarios      ─┘                                  └─ Google TTS / ElevenLabs / gTTS
                         app/container.py nối port ↔ adapter
```

- `app/domain` — entity, quy tắc hội thoại, port (interface). Không import FastAPI/SQLAlchemy/Twilio.
- `app/application` — use case: `VoiceCallService`, `CallService`, `ScenarioService`, `SpeechService`.
- `app/adapters/inbound` — HTTP (routes, schemas, TwiML, auth) và CLI.
- `app/adapters/outbound` — persistence, LLM, TTS.
- `app/container.py` — composition root, nơi duy nhất biết cả hai phía.

Chi tiết: [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md).

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
python -m app.adapters.inbound.cli.seed_scenarios
```

### 7. Chạy server

```bash
uvicorn app.adapters.inbound.api.main:app --reload
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
docker-compose exec api python -m app.adapters.inbound.cli.seed_scenarios
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

Mọi endpoint `/api/*` yêu cầu quyền admin (`Authorization: Bearer <session_token>` của admin, hoặc header `X-Internal-Token`).

### Voice Endpoints (webhook Twilio, có kiểm tra chữ ký `X-Twilio-Signature`)

| Endpoint | Mô tả |
|---|---|
| `POST /voice/incoming` | Cuộc gọi đến: chọn kịch bản theo số được gọi, đọc lời chào |
| `POST /voice/process-speech` | Mỗi lượt khách nói: tra FAQ, gọi AI, trả lời / kết thúc / chuyển máy |
| `POST /voice/status` | Status callback: `queued`, `ringing`, `in-progress`, `completed`, `busy`, `no-answer`, `failed`, `canceled` (trạng thái lạ bị bỏ qua) |

### Cuộc gọi

```bash
GET    /api/calls?status=completed&outcome=transferred&from_number=+84...
GET    /api/calls/{call_id}          # kèm toàn bộ hội thoại
DELETE /api/calls/{call_id}
GET    /api/analytics/statistics?days=30
```

`outcome`: `resolved` (bot giải quyết xong), `transferred` (đã chuyển máy), `callback_requested` (cần gọi lại), `abandoned` (khách cúp máy giữa chừng).

### Kịch bản

```bash
GET    /api/scenarios
POST   /api/scenarios/seed           # nạp 6 kịch bản + FAQ mẫu (idempotent)
POST   /api/scenarios
{
  "name": "hotline_vip",
  "description": "Tổng đài khách VIP",
  "greeting": "Xin chào quý khách VIP...",
  "prompt": "Bạn là chuyên viên chăm sóc khách VIP...",
  "phone_number": "+842873000000",   # cuộc gọi tới số này dùng kịch bản này
  "is_default": 0,                   # 1 = dùng cho số chưa gắn kịch bản (chỉ có một mặc định)
  "is_active": 1
}
GET | PUT | DELETE /api/scenarios/{scenario_id}
```

Thứ tự chọn kịch bản: theo `phone_number` → kịch bản `is_default` → `VOICE_SCENARIO` → kịch bản dựng sẵn.

### Kho kiến thức (FAQ)

```bash
GET    /api/knowledge?category=giao hàng&scenario=sales
GET    /api/knowledge/search?q=phí ship bao nhiêu&scenario=customer_support
POST   /api/knowledge
{
  "question": "Có bán sỉ không?",
  "answer": "Có, đơn từ 50 sản phẩm được giá sỉ.",
  "keywords": "sỉ, buôn, số lượng lớn",
  "category": "bán hàng",
  "scenario": null                   # null = dùng cho mọi kịch bản
}
GET | PUT | DELETE /api/knowledge/{entry_id}
```

Mỗi lượt khách nói, bot lấy `KNOWLEDGE_TOP_K` mục khớp nhất (theo câu hỏi + keywords) đưa vào prompt. Câu trả lời trong FAQ nên viết sẵn theo văn nói, vì bot sẽ đọc gần như nguyên văn.

### Khách hàng

```bash
GET /api/customers/{phone_number}
PUT /api/customers/{phone_number}   {"name": "Chị Lan", "notes": "Khách VIP, ưu tiên"}
```

### Giả lập & kiểm thử

```bash
# Bắt đầu cuộc gọi giả lập (trả về call_sid + lời chào)
POST /api/simulate {"scenario": "sales"}
# Các lượt tiếp theo
POST /api/simulate {"call_sid": "SIM-...", "message": "phí ship bao nhiêu"}
# -> {"reply": "...", "listening": true, "ended": false, "transfer_to": null, "knowledge": ["Phí giao hàng là bao nhiêu?"]}

POST /api/test/tts?text=Xin chào&language=vi-VN
```

## 🎭 Kịch bản có sẵn

| Kịch bản | Dùng cho |
|---|---|
| `customer_support` (mặc định) | Câu hỏi chung: đơn hàng, giao hàng, đổi trả, thanh toán, bảo hành, khuyến mãi, tài khoản |
| `order_support` | Kiểm tra / hủy / sửa đơn, đổi trả, hoàn tiền |
| `sales` | Tư vấn sản phẩm, khuyến mãi, thanh toán |
| `complaint` | Tiếp nhận khiếu nại, thu thập thông tin, chuyển nhân viên khi cần |
| `appointment` | Đặt lịch hẹn: họ tên, số điện thoại, dịch vụ, thời gian |
| `survey` | Khảo sát mức độ hài lòng |

Mọi kịch bản đều dùng chung bộ quy tắc gọi thoại: trả lời ngắn theo văn nói, chỉ dùng thông tin trong FAQ / thông tin công ty, không hỏi OTP hay mật khẩu, tự kết thúc cuộc gọi khi khách chào tạm biệt, và chuyển nhân viên khi không xử lý được.

> ⚠️ 17 mục FAQ mẫu (phí ship 300 nghìn, đổi trả 7 ngày, hoàn tiền 3–7 ngày, ...) chỉ là **dữ liệu minh hoạ**. Hãy sửa lại theo chính sách thật qua `/api/knowledge` trước khi chạy thật.

## 💾 Database Schema

Bảng được tạo tự động khi khởi động. Cột mới được tự thêm vào DB cũ (`ALTER TABLE ... ADD COLUMN`).

| Bảng | Cột chính |
|---|---|
| `calls` | call_sid, from_number, to_number, status, duration, start/end_time, notes, **scenario**, **outcome** |
| `messages` | call_id, role, content, timestamp, processing_time |
| `scenarios` | name, description, prompt, **greeting**, **phone_number**, **is_default**, is_active |
| `knowledge_entries` | question, answer, keywords, category, scenario, is_active |
| `customers` | phone_number, name, email, notes, total_calls, last_call_date |

## 🔊 Text-to-Speech

Thứ tự thử: Google Cloud TTS (`GOOGLE_TTS_ENABLED`, `GOOGLE_APPLICATION_CREDENTIALS`) → ElevenLabs (`ELEVENLABS_API_KEY`) → gTTS. Hiện TTS chỉ dùng cho `/api/test/tts`; câu trả lời trong cuộc gọi do Twilio `<Say>` đọc.

## 🤖 AI

| Biến | Mặc định | Ghi chú |
|---|---|---|
| `ANTHROPIC_API_KEY` | | Provider chính |
| `ANTHROPIC_MODEL` | `claude-opus-5-5` | |
| `ANTHROPIC_EFFORT` | `low` | Giảm độ trễ cho cuộc gọi thoại; tăng lên `medium` nếu cần trả lời kỹ hơn |
| `ANTHROPIC_MAX_TOKENS` | `2048` | Tính cả token suy luận (thinking) |
| `ANTHROPIC_REFUSAL_FALLBACK` | `true` | Server-side fallback (`fallbacks: "default"`) khi model từ chối; tắt nếu dùng model không hỗ trợ (vd Haiku) |
| `OPENAI_API_KEY`, `OPENAI_MODEL` | `gpt-4-turbo-preview` | Dự phòng khi Claude lỗi |
| `LLM_TIMEOUT_SECONDS` | `10` | Twilio chờ webhook tối đa 15 giây |

Yêu cầu Python ≥ 3.10 (`anthropic` 1.x).

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
