# Ollama Gateway

## Cấu trúc dự án

```
ollama-gateway/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── server/
│   │   └── server.go        # WebSocket server + routing logic
│   ├── search/
│   │   └── engine.go        # Full-text search (TF-IDF, in-memory)
│   ├── ollama/
│   │   └── client.go        # Ollama API client (streaming, vision)
│   └── preset/
│       └── loader.go        # Load preset Q&A từ folder JSON
├── presets/
│   └── company_faq.json     # Câu hỏi mặc định (thêm file .json tùy ý)
├── go.mod
└── README.md
```

## Cài đặt

```bash
go mod tidy
go run cmd/main.go
```

## Các flag

| Flag | Mặc định | Mô tả |
|------|----------|-------|
| `-addr` | `:8080` | Địa chỉ WebSocket server |
| `-presets` | `./presets` | Folder chứa file Q&A JSON |
| `-ollama` | `http://localhost:11434` | URL Ollama |
| `-model` | `llama3.2` | Tên model Ollama |

```bash
go run cmd/main.go \
  -addr :9000 \
  -presets ./my-presets \
  -ollama http://localhost:11434 \
  -model llava  # dùng llava cho vision (ảnh)
```

## Luồng xử lý

```
Client WS ──► Server
                │
                ├─ Full-Text Search (presets folder)
                │       │
                │   Score >= 0.5? ──YES──► Trả preset answer ngay
                │       │
                │       NO
                │       │
                └──────►└─► Ollama API (streaming tokens)
                                │
                            Có ảnh? ──► Vision model (llava)
```

## Thêm preset câu hỏi

Tạo file `.json` mới trong folder `presets/`:

```json
{
  "category": "my_topic",
  "items": [
    {
      "id": "unique_id",
      "keywords": ["từ khóa 1", "từ khóa 2", "keyword"],
      "question": "Câu hỏi mẫu",
      "answer": "Câu trả lời đầy đủ ở đây",
      "image_url": "https://example.com/image.png"
    }
  ]
}
```

## WebSocket Protocol

**Client → Server:**
```json
{
  "text": "Câu hỏi của user",
  "images": ["base64_string_hoặc_đường_dẫn_file"]
}
```

**Server → Client (streaming):**
```json
{ "type": "preset",  "content": "Câu trả lời preset", "source": "preset" }
{ "type": "token",   "content": "từng token...",       "source": "ollama" }
{ "type": "done",    "content": "",                    "source": "ollama" }
{ "type": "error",   "content": "Lỗi gì đó",          "source": "ollama" }
```

## Test

Mở browser: http://localhost:8080

Hoặc dùng websocat:
```bash
websocat ws://localhost:8080/ws
{"text": "xin chào"}
```

## Điều chỉnh ngưỡng search

Trong `internal/server/server.go`:
```go
const searchThreshold = 0.5  // 0.0-1.0, cao hơn = khắt khe hơn
```
