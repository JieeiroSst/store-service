# lottery-service

Cào & kiểm tra kết quả xổ số kiến thiết Việt Nam (Miền Nam, Miền Trung, Miền Bắc).

Kiến trúc **Hexagonal (Ports & Adapters)**:

```
app/
├── domain/                 # Lõi nghiệp vụ — KHÔNG import FastAPI/Motor
│   ├── models.py           # Entity: LotteryResult, PrizeMatch, WinningCheckResult...
│   ├── prize_rules.py      # Quy tắc dò thưởng 3 miền (pure functions, có strategy theo miền)
│   └── ports/
│       ├── result_repository.py   # Port outbound: lưu/đọc kết quả
│       └── scraper_port.py        # Port outbound: nguồn cào dữ liệu
├── application/             # Use case, điều phối domain qua port
│   ├── scrape_results.py    # Cào 1 ngày, cả 3 miền, lưu qua repo
│   ├── backfill.py          # Cào theo khoảng ngày, resume-safe
│   └── check_winning.py     # Dò thưởng theo (ngày, tỉnh, số vé)
├── adapters/
│   ├── inbound/api/         # FastAPI routers = inbound adapter
│   ├── outbound/mongo/      # MongoDB adapter (Motor) implement result_repository
│   ├── outbound/scrapers/   # HTTP scraper adapters (1 file / miền) implement scraper_port
│   └── scheduler/           # APScheduler chạy use case scrape hằng ngày
├── config.py                 # pydantic-settings
├── container.py               # DI: nối port <-> adapter
└── cli.py                     # CLI backfill / scrape thủ công
tests/
```

Nguyên tắc phụ thuộc: `adapters → application → domain`. Domain/Application không
biết gì về FastAPI hay Mongo — chỉ biết `ResultRepositoryPort` và `ScraperPort`.
`container.py` là nơi duy nhất "nối dây" adapter thật vào use case.

## Cài đặt & chạy local

Yêu cầu Python 3.11+ và MongoDB đang chạy (local hoặc Docker).

```bash
python3.11 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

cp .env.example .env   # chỉnh MONGO_URI nếu cần

uvicorn app.adapters.inbound.api.main:app --reload
```

API mặc định chạy ở `http://localhost:8000`, Swagger UI tại `/docs`.

## Chạy bằng Docker

```bash
docker compose up -d --build
```

Lệnh trên khởi động cả MongoDB và lottery-service (API ở cổng 8000, scheduler
tự bật qua biến môi trường `SCHEDULER_ENABLED=true` trong `docker-compose.yml`).

## Triển khai lên Kubernetes

Chart nằm ở [`chart/lottery-service`](../chart/lottery-service) (Deployment +
Service + HPA + Ingress tuỳ chọn), theo đúng convention chung của repo:

- Cấu hình runtime (`MONGO_URI`, `MONGO_DB`, `SCHEDULER_ENABLED`, ...) không
  truyền qua `env:` trong Deployment, mà được mount thành file `/app/.env`
  từ Secret dùng chung `service-env-files` (key `lottery-service.env`, định
  nghĩa ở [`chart/templates/env-files-secret.yaml`](../chart/templates/env-files-secret.yaml)).
  `pydantic-settings` trong `app/config.py` tự đọc file `.env` này khi container start.
- `MONGO_URI` trỏ tới **MongoDB dùng chung trong cụm** qua Service `mongo-svc`
  (chart [`chart/mongo`](../chart/mongo)), cổng **80** — đúng convention các
  service khác đang dùng cho hạ tầng chung (`redis-svc:80`, `mysql-svc:80`),
  vì Service của các chart hạ tầng này luôn expose ở port 80 rồi mới
  `targetPort` sang cổng thật của container (Mongo là 27017):

  ```
  MONGO_URI=mongodb://mongo-svc:80
  MONGO_DB=lottery
  ```

  Không cần tạo trước database `lottery` trong cụm Mongo dùng chung: MongoDB
  tạo database/collection theo kiểu lazy (chỉ khi có ghi/tạo index đầu tiên),
  và `container.startup()` đã gọi `ensure_indexes()` ngay lúc app khởi động
  (xem `app/adapters/outbound/mongo/repository.py`) nên database + collection
  `lottery_results` cùng 2 index tự sinh ra ngay lần deploy đầu tiên — không
  cần bước provision DB thủ công nào thêm.
- Readiness/liveness probe dùng endpoint `GET /health` có sẵn.
- Autoscaling **tắt mặc định** (`lottery.autoscaling.enabled: false`, 1
  replica) — Deployment chỉ phục vụ API, không tự cào dữ liệu (xem CronJob bên
  dưới), nên bật `lottery.autoscaling.enabled=true` thoải mái nếu muốn scale
  theo tải API mà không lo cào trùng.

### Cào dữ liệu hằng ngày bằng k8s CronJob

Việc cào tự động hằng ngày chạy bằng **CronJob** riêng
([`templates/scrape-cronjob.yaml`](../chart/lottery-service/templates/scrape-cronjob.yaml)),
KHÔNG phải scheduler nội bộ (APScheduler) chạy trong process API:

- Lịch: `schedule: "0 12 * * *"` (giờ UTC của cluster) = **19:00 (7h tối) giờ
  Việt Nam** mỗi ngày (`Asia/Ho_Chi_Minh` = UTC+7, không có giờ mùa hè nên quy
  đổi cố định). Đổi ở `lottery.scrapeCronJob.schedule` trong `values.yaml`.
- Lệnh chạy: `python -m app.cli scrape` — không truyền `--date` sẽ tự lấy
  **ngày hôm nay theo giờ Việt Nam** (`app/cli.py`, dùng `zoneinfo`), nên
  CronJob không cần biết/tính ngày.
- `concurrencyPolicy: Forbid` — nếu lần chạy trước chưa xong (vd nguồn chậm)
  thì bỏ qua lần kích hoạt tiếp theo thay vì chạy chồng.
- Dùng chung image + `lottery-service.env` (mount qua Secret
  `service-env-files`) với Deployment API, nên không cần cấu hình Mongo riêng.
- Vì CronJob đã đảm nhiệm việc này, `SCHEDULER_ENABLED=false` trong
  `lottery-service.env` (tắt APScheduler nội bộ) — APScheduler nội bộ
  (`app/adapters/scheduler/daily_job.py`, bật qua `SCHEDULER_ENABLED=true`)
  chỉ còn dùng cho docker-compose/chạy local, nơi không có k8s CronJob.

Xem log / trigger thủ công:

```bash
kubectl get cronjob lottery-service-scrape
kubectl create job --from=cronjob/lottery-service-scrape lottery-scrape-manual-$(date +%s)
kubectl logs -l job-name=<tên job vừa tạo> -f
```

CI/CD: [`.github/workflows/lottery-service-cicd.yml`](../.github/workflows/lottery-service-cicd.yml)
— build & `pytest`, build/push image đa kiến trúc lên
`ghcr.io/jieeirosst/lottery-service`, rồi `helm template ./chart/lottery-service | kubectl apply`
trên self-hosted runner (giống các service khác trong repo, không dùng
`helm install` để tránh xung đột ownership với umbrella release).

Deploy thủ công (không qua CI):

```bash
helm template ./chart/lottery-service --namespace default \
  --set lottery.image.tag=<tag> \
  | kubectl apply -n default -f -
kubectl rollout status deployment/lottery-service-deployment -n default
```

## Gọi API — ví dụ curl

Tất cả ví dụ bên dưới cũng có sẵn ở [`curl-examples.sh`](curl-examples.sh), chạy được luôn:

```bash
./curl-examples.sh                              # chạy hết, vào http://localhost:8000
BASE_URL=http://localhost:8010 ./curl-examples.sh  # đổi host/port
./curl-examples.sh check_won                    # chỉ chạy 1 ví dụ (tên hàm trong file)
```

### Kiểm tra trúng thưởng

```bash
curl -X POST http://localhost:8000/api/v1/check \
  -H "Content-Type: application/json" \
  -d '{"date": "2026-08-26", "province": "Đồng Nai", "ticket_number": "145917"}'
```

Response:

```json
{
  "won": true,
  "prizes_won": ["dac_biet"],
  "detail": [
    {"prize": "dac_biet", "matched": "145917", "rule": "trùng đủ 6 số giải đặc biệt"}
  ],
  "result_found": true,
  "highest_prize": "dac_biet"
}
```

`province` chấp nhận cả tên tỉnh có dấu ("Đồng Nai") lẫn `province_code` đã
chuẩn hoá ("dongnai") — hệ thống tự bỏ dấu/khoảng trắng để so khớp.

Nếu chưa có kết quả cho (ngày, tỉnh) đó, API trả **HTTP 200** với
`result_found: false` (đây là một kết quả nghiệp vụ hợp lệ — "chưa có dữ
liệu" — không phải lỗi hệ thống, nên không dùng mã lỗi 4xx/5xx).

### Lấy nguyên bản kết quả 1 tỉnh/đài

```bash
curl "http://localhost:8000/api/v1/results?date=2026-08-26&province=Đồng%20Nai"
```

Trả **HTTP 404** nếu không có dữ liệu (đây là truy vấn "lấy 1 resource cụ
thể" nên 404 hợp lý hơn, khác với `/check` ở trên).

### Cào thủ công 1 ngày (3 miền)

```bash
curl -X POST http://localhost:8000/api/v1/scrape -H "Content-Type: application/json" \
  -d '{"date": "2026-08-26"}'
```

### Backfill qua API

```bash
curl -X POST http://localhost:8000/api/v1/backfill -H "Content-Type: application/json" \
  -d '{"start": "2026-01-01", "end": "2026-01-31", "force": false}'
```

## Backfill lịch sử bằng CLI

```bash
python -m app.cli backfill --start 2026-01-01 --end 2026-01-31
python -m app.cli backfill --start 2026-01-01 --end 2026-01-31 --force   # cào lại kể cả ngày đã có
python -m app.cli scrape --date 2026-08-26
```

Backfill **resume-safe**: với mỗi ngày, mỗi miền đã có dữ liệu trong MongoDB sẽ
tự động bị bỏ qua (trừ khi `--force`), nên có thể dừng giữa chừng và chạy lại
mà không tạo dữ liệu trùng hay tốn thời gian cào lại từ đầu.

## Bật scheduler chạy hằng ngày (docker-compose / local)

Trên **Kubernetes**, việc cào hằng ngày do k8s CronJob đảm nhiệm (xem mục
["Cào dữ liệu hằng ngày bằng k8s CronJob"](#cào-dữ-liệu-hằng-ngày-bằng-k8s-cronjob)
ở trên) — không cần bật scheduler nội bộ.

Với docker-compose hoặc chạy local (không có k8s), đặt `SCHEDULER_ENABLED=true`
trong `.env` (docker-compose.yml đã bật sẵn). Scheduler dùng APScheduler, chạy
vào các giờ cấu hình ở `SCHEDULER_RUN_HOURS` (mặc định `[18, 19, 21]` giờ Việt
Nam — nhiều mốc để phòng trường hợp nguồn cập nhật muộn hoặc lần chạy đầu lỗi mạng).

## Mô hình dữ liệu MongoDB

Collection `lottery_results`, mỗi document = kết quả 1 tỉnh/đài trong 1 ngày:

```json
{
  "date": "2026-08-26",
  "region": "mien-nam",
  "province": "Đồng Nai",
  "province_code": "dongnai",
  "prizes": {
    "dac_biet": ["145917"],
    "giai_1": ["98208"],
    "giai_8": ["70"]
  },
  "source": "minhchinh.com",
  "scraped_at": "2026-08-26T11:05:00Z"
}
```

- Unique index trên `(date, region, province_code)` — chống trùng lặp, ghi
  theo cơ chế **upsert** nên chạy lại không tạo bản ghi thừa.
- Index phụ trên `date` để truy vấn theo ngày nhanh.

## Quy tắc dò thưởng (`app/domain/prize_rules.py`)

Bộ quy tắc tách theo `RegionRuleSet` (strategy) cho từng miền, dễ chỉnh khi
công ty xổ số đổi thể lệ mà không đụng vào logic dò chung:

- **Miền Nam / Miền Trung** (vé 6 chữ số): ĐB khớp đủ 6 số; G1..G8 so khớp
  N chữ số CUỐI của vé với con số của giải (G8=2, G7=3, G6=4, G5=4, G4=5,
  G3=5, G2=5, G1=5 chữ số). **Giải phụ đặc biệt**: 5 số cuối trùng ĐB, riêng
  số đầu tiên sai khác. **Giải khuyến khích** (bật/tắt qua
  `ENABLE_KHUYEN_KHICH_RULE`): sai đúng 1 chữ số ở bất kỳ vị trí nào so với ĐB.
- **Miền Bắc** (vé 5 chữ số): ĐB, G1..G7 theo cùng nguyên tắc so đuôi (G7=2,
  G6=3, G5=4, G4=4, G3=5, G2=5, G1=5, ĐB=5 chữ số); không có phụ ĐB/khuyến khích.

Một vé có thể trúng nhiều giải cùng lúc — API liệt kê tất cả trong
`prizes_won`/`detail`, đồng thời trả `highest_prize` là giải cao nhất.

## Cào dữ liệu (scraper)

Mỗi miền là 1 adapter độc lập implement `ScraperPort.fetch(day) -> list[LotteryResult]`
(`app/adapters/outbound/scrapers/{mien_nam,mien_trung,mien_bac}.py`). Parser
(`base.py: parse_province_blocks`) nhận diện giải qua **nhãn text** của từng
hàng (vd "G.ĐB", "100N", "2TỶ") map qua `LABEL_MAP_*`, thay vì phụ thuộc vào
tên class CSS cụ thể — chịu được thay đổi giao diện miễn nhãn còn giữ ý nghĩa.
Việc parse HTML tách hoàn toàn khỏi việc gọi mạng (`parse_html(html, day, source)`
là pure function) nên test được 100% offline bằng HTML mẫu ở `tests/fixtures/`.

HTTP client dùng `httpx.AsyncClient`, có retry + backoff
(`HTTP_MAX_RETRIES`, `HTTP_RETRY_BACKOFF_SECONDS`), User-Agent giống trình
duyệt thật, và delay giữa các ngày khi backfill (`BACKFILL_DELAY_SECONDS`) để
lịch sự với server nguồn.

> **Lưu ý:** cấu trúc HTML cụ thể trong `parse_province_blocks` là một khuôn
> mẫu hợp lý cho một trang tổng hợp theo ngày (mỗi tỉnh là 1 khối
> `data-tinh`/`data-ma`, mỗi hàng `<tr><td>nhãn</td><td>số</td></tr>`); vì
> giao diện trang nguồn thật có thể thay đổi theo thời gian, hãy đối chiếu HTML
> thực tế và điều chỉnh lại đúng mỗi `parse_province_blocks`/`build_url` — phần
> còn lại của hệ thống (domain, application, API, DB) không cần thay đổi gì nhờ
> kiến trúc hexagonal.

## Chạy test

```bash
pip install -r requirements.txt
pytest -q
```

Bao gồm:
- `test_prize_rules.py` — quy tắc dò thưởng: trúng ĐB, trúng phụ ĐB, trúng
  khuyến khích (bật/tắt), trúng giải thấp, trúng nhiều giải cùng lúc, không
  trúng, và bộ quy tắc riêng của Miền Bắc.
- `test_scrapers.py` — parser từng miền bằng HTML mẫu (`tests/fixtures/`), không cần mạng.
- `test_repository.py` — `MongoResultRepository` bằng `mongomock-motor` (giả lập Motor).
- `test_use_cases.py` — use case scrape/backfill/check bằng fake scraper & fake repository.
- `test_check_winning_api.py` — API `/api/v1/check` qua FastAPI `TestClient` + fake repository.

## Ghi chú pháp lý

Dữ liệu kết quả xổ số là thông tin công khai do các công ty xổ số kiến thiết
công bố. Hệ thống này chỉ tổng hợp lại để tra cứu, có đặt delay + retry hợp lý
khi cào để không gây tải cho server nguồn, và **không phục vụ mục đích cá cược,
số đề hay bất kỳ hình thức cờ bạc trái phép nào**.
