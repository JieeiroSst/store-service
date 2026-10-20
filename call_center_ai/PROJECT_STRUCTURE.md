# 📁 Project Structure — Hexagonal Architecture

```
call_center_ai/
├── app/
│   ├── config.py                       # Settings (pydantic-settings, đọc .env / env vars)
│   ├── container.py                    # Composition root: nối port ↔ adapter, build use case
│   │
│   ├── domain/                         # Lõi nghiệp vụ, không phụ thuộc framework
│   │   ├── models.py                   # Call, Message, Scenario, KnowledgeEntry, Customer, CallOutcome, VoiceReply
│   │   ├── conversation.py             # Câu thoại, ý định (kết thúc/gặp nhân viên), [[END]]/[[TRANSFER]], prompt builder
│   │   ├── knowledge.py                # Chuẩn hoá tiếng Việt + chấm điểm FAQ
│   │   ├── playbook.py                 # 6 kịch bản dựng sẵn + FAQ mẫu
│   │   ├── errors.py                   # NotFoundError, AlreadyExistsError
│   │   └── ports/
│   │       ├── repositories.py         # Call/Message/Scenario/Knowledge/CustomerRepositoryPort
│   │       ├── llm.py                  # LLMPort
│   │       └── tts.py                  # TTSPort
│   │
│   ├── application/                    # Use cases, chỉ dùng domain + ports
│   │   ├── voice.py                    # VoiceCallService: start_call, handle_speech, update_status, handoff
│   │   ├── calls.py                    # CallService: list/detail/delete, statistics, customer
│   │   ├── scenarios.py                # ScenarioResolver + ScenarioService: CRUD, seed_defaults
│   │   ├── knowledge.py                # KnowledgeService: CRUD, search, seed_defaults
│   │   └── speech.py                   # SpeechService: TTS
│   │
│   └── adapters/
│       ├── inbound/
│       │   ├── api/
│       │   │   ├── main.py             # FastAPI app, lifespan, exception handlers
│       │   │   ├── routes/voice.py     # Twilio webhooks /voice/*
│       │   │   ├── routes/admin.py     # REST /api/*: calls, scenarios, knowledge, customers, simulate
│       │   │   ├── schemas.py          # Pydantic request/response
│       │   │   ├── twiml.py            # VoiceReply → TwiML (Gather/Say/Dial/Hangup)
│       │   │   ├── dependencies.py     # FastAPI Depends → use case từ container
│       │   │   ├── authn.py, auth_deps.py   # Xác thực qua user_service / internal token
│       │   │   └── twilio_auth.py      # Kiểm tra chữ ký X-Twilio-Signature
│       │   └── cli/seed_scenarios.py   # Seed kịch bản + FAQ mẫu
│       └── outbound/
│           ├── persistence/            # SQLAlchemy: orm.py, db.py (+ tự thêm cột thiếu), repositories.py
│           ├── llm/                    # AnthropicLLM, OpenAILLM, FallbackLLM
│           └── tts/                    # GoogleTTS, ElevenLabsTTS, GTTS, FallbackTTS
│
├── tests/                              # Unit + API tests, fake adapters in-memory (fakes.py)
├── scripts/smoke_test.py               # Kiểm tra kết nối DB/LLM/TTS thật
├── Dockerfile, docker-compose.yml, requirements.txt, .env.example
```

## Quy tắc phụ thuộc

```
adapters/inbound ──▶ application ──▶ domain ◀── adapters/outbound
                          ▲                         ▲
                          └──── container.py ───────┘
```

- `domain` không import gì ngoài stdlib.
- `application` chỉ import `domain` (kể cả `domain.ports`).
- Adapter implement port (outbound) hoặc gọi use case (inbound). Route không chứa logic nghiệp vụ.
- Chỉ `container.py` (và `main.py` / CLI qua container) biết adapter cụ thể nào được dùng.

## Luồng một cuộc gọi

1. `POST /voice/incoming` → `VoiceCallService.start_call`
   - tạo/cập nhật customer (`total_calls`, `last_call_date`);
   - `ScenarioResolver.for_number(To)`: kịch bản gắn với số được gọi → kịch bản `is_default` → `VOICE_SCENARIO` → kịch bản dựng sẵn;
   - lưu call (`scenario`) và lời chào của kịch bản → `VoiceReply` → `twiml.render`.
2. `POST /voice/process-speech` → `handle_speech`
   - lưu câu khách nói; nếu khách đòi gặp nhân viên (`wants_human`) hoặc vượt `VOICE_MAX_TURNS` → `_handoff`;
   - dựng history từ `messages`, tìm `KNOWLEDGE_TOP_K` FAQ khớp với 2 câu gần nhất của khách;
   - `build_system_prompt` = prompt kịch bản + `CALL_RULES` + thông tin công ty + FAQ + hồ sơ khách → `LLMPort`;
   - `parse_directive`: `[[TRANSFER]]` → `_handoff`; `[[END]]` hoặc khách nói tạm biệt → hoàn tất, outcome `resolved`; còn lại → tiếp tục nghe.
   - `_handoff`: có `AGENT_PHONE_NUMBER` → `<Dial>`, outcome `transferred`; không có → hứa gọi lại, outcome `callback_requested`.
3. `POST /voice/status` → map trạng thái Twilio → `Call.apply_provider_status`. Trạng thái trung gian không ghi đè cuộc gọi đã kết thúc; cuộc gọi `completed` chưa có outcome thì thành `abandoned`.

History hội thoại được dựng lại từ DB mỗi lượt (không còn dict global trong process), nên chạy nhiều replica/restart vẫn giữ ngữ cảnh.

`POST /api/simulate` dùng đúng `VoiceCallService`, chỉ khác kênh vào (JSON thay vì Twilio), nên kết quả giả lập giống hệt cuộc gọi thật.

## Thêm adapter mới

Ví dụ thêm Gemini: tạo `adapters/outbound/llm/gemini_llm.py` implement `LLMPort`, rồi thêm vào `build_llm_providers` trong `container.py`. Domain và use case không đổi.

## Chạy

```bash
pip install -r requirements.txt
python -m app.adapters.inbound.cli.seed_scenarios
uvicorn app.adapters.inbound.api.main:app --reload
python -m unittest discover -s tests -t .
```

`DATABASE_URL` (tuỳ chọn) ghi đè cấu hình `DB_*`, ví dụ `sqlite:///local.db` khi dev.
