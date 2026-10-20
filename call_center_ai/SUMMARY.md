# 📋 Call Center AI — Tổng quan

## Hệ thống làm gì

1. Khách gọi vào số Twilio → `/voice/incoming` chọn kịch bản theo số được gọi, đọc lời chào và nghe khách nói.
2. Mỗi câu khách nói → `/voice/process-speech`:
   - khách đòi gặp người thật → chuyển máy ngay (hoặc ghi nhận gọi lại nếu không có ai trực);
   - còn lại: tìm các FAQ khớp nhất, ghép với prompt kịch bản, thông tin công ty và hồ sơ khách, rồi gọi Claude;
   - Claude trả lời theo văn nói và có thể đánh dấu `[[END]]` (kết thúc) hoặc `[[TRANSFER]]` (chuyển nhân viên). Bot xoá các dấu này trước khi đọc.
3. Twilio gửi status callback → cập nhật trạng thái, thời lượng và kết quả (`resolved`, `transferred`, `callback_requested`, `abandoned`).
4. Admin quản lý kịch bản, FAQ, khách hàng, xem lịch sử, thống kê và giả lập hội thoại qua `/api/*`.

## Các quyết định thiết kế

| Quyết định | Lý do |
|---|---|
| Hexagonal (domain / application / adapters) | Đổi DB, LLM, TTS hay kênh vào (Twilio, CLI, simulate) mà không động tới nghiệp vụ; test use case bằng fake adapters |
| Lịch sử hội thoại dựng lại từ bảng `messages` | Chạy nhiều replica, restart không mất ngữ cảnh |
| Bot chỉ trả lời từ FAQ + thông tin công ty | Tránh bịa chính sách, giá, thời hạn |
| Tìm FAQ bằng từ khoá tiếng Việt (bỏ dấu, stopwords, bigram) | Đủ tốt cho vài trăm mục, không cần vector DB; thay được bằng adapter khác sau này |
| Claude `claude-opus-5-5`, effort `low`, timeout 10s | Twilio chỉ chờ webhook 15 giây; effort thấp giữ độ trễ thấp cho hội thoại ngắn |
| Server-side refusal fallback | Khi model từ chối, API tự chạy lại trên model dự phòng trong cùng request |
| Chuyển máy bằng TwiML `<Dial>` | Không cần hệ thống ACD riêng; để trống số nhân viên thì chuyển thành yêu cầu gọi lại |
| Tự thêm cột thiếu khi khởi động | DB MySQL đang chạy nhận cột mới mà không cần công cụ migration |

## Giới hạn hiện tại

- Bot không tra cứu được hệ thống đơn hàng thật; câu hỏi về một đơn cụ thể sẽ được chuyển nhân viên.
- FAQ mẫu chỉ để minh hoạ, cần thay bằng chính sách thật.
- `OPENAI_MODEL` mặc định (`gpt-4-turbo-preview`) và SDK `openai==1.3.5` chưa được nâng cấp.
- TTS (Google/ElevenLabs/gTTS) mới dùng cho `/api/test/tts`; trong cuộc gọi, Twilio `<Say>` đọc câu trả lời.
