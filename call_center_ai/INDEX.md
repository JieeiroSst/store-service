# 📚 Call Center AI — Mục lục tài liệu

| Tài liệu | Nội dung | Khi nào đọc |
|---|---|---|
| [START_HERE.txt](START_HERE.txt) | Lệnh chạy nhanh, checklist trước khi chạy thật | Lần đầu mở project |
| [QUICKSTART.md](QUICKSTART.md) | Setup từng bước (venv, MySQL, Twilio, ngrok) | Muốn chạy thử trong 10 phút |
| [README.md](README.md) | Tính năng, biến môi trường, toàn bộ API, kịch bản, schema | Tra cứu hằng ngày |
| [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) | Kiến trúc hexagonal, quy tắc phụ thuộc, luồng cuộc gọi | Trước khi sửa code |
| [DEPLOYMENT.md](DEPLOYMENT.md) | Triển khai production (supervisor, nginx, Docker) | Khi deploy |
| [SUMMARY.md](SUMMARY.md) | Tổng quan chức năng và các quyết định thiết kế | Giới thiệu project cho người mới |
| [FILE_MANIFEST.txt](FILE_MANIFEST.txt) | Danh sách file và vai trò | Tìm file nhanh |

## Tra cứu theo nhu cầu

| Tôi muốn... | Xem |
|---|---|
| Thêm / sửa câu trả lời FAQ | README → Kho kiến thức; `app/domain/playbook.py` cho dữ liệu mẫu |
| Tạo kịch bản mới cho một số tổng đài | README → Kịch bản (`phone_number`, `is_default`) |
| Đổi cách bot nói (độ dài, xưng hô, quy tắc) | `app/domain/conversation.py` (`CALL_RULES`, `build_system_prompt`) |
| Chuyển máy cho nhân viên | `.env` → `AGENT_PHONE_NUMBER`; logic ở `VoiceCallService._handoff` |
| Thử kịch bản không cần gọi điện | `POST /api/simulate` |
| Đổi model AI | `.env` → `ANTHROPIC_MODEL`, `ANTHROPIC_EFFORT`, `OPENAI_MODEL` |
| Thêm provider LLM / TTS | PROJECT_STRUCTURE → Thêm adapter mới |
| Xem cuộc gọi cần gọi lại | `GET /api/calls?outcome=callback_requested` |
| Chạy test | `python -m unittest discover -s tests -t .` |
