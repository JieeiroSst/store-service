FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ekyc-service ./cmd

# The final stage is python:3.10-slim (glibc), not alpine like every other
# service in this repo: the CARD_READER_PROVIDER=python/
# FACE_ANALYZER_PROVIDER=python adapters (internal/adapter/secondary/
# pyextract) shell out to python/extract_id.py, which needs dlib
# (face_recognition) and torch (EasyOCR). Both ship prebuilt manylinux
# wheels but not musllinux/alpine-compatible ones for the versions this
# pins, so alpine would force a from-source dlib compile (slow, and prone
# to failing in a minimal build environment). This image is much larger
# as a result (~1-2GB) - an explicit, deliberate tradeoff for the
# accuracy those providers give when enabled; the default providers
# ("native", see config.ProvidersConfig) don't use Python at all and work
# fine even if it's misconfigured or missing.
FROM python:3.10-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

COPY python/requirements.txt ./python/requirements.txt
# dlib-bin ships prebuilt wheels (see requirements.txt's comment) so this
# does not compile anything from source; face_recognition is installed
# with --no-deps against it so pip's resolver doesn't also pull the real
# (source-only) `dlib` package as a transitive dependency.
RUN pip install --no-cache-dir dlib-bin==20.0.1 face-recognition-models==0.3.0 \
    && pip install --no-cache-dir --no-deps face_recognition==1.3.0 \
    && pip install --no-cache-dir click==8.1.7 numpy==1.26.4 pillow==10.4.0 easyocr==1.7.2

COPY python/extract_id.py ./python/extract_id.py

COPY --from=builder /app/ekyc-service .
COPY --from=builder /app/database.sql .

EXPOSE 8087

ENTRYPOINT ["./ekyc-service"]
