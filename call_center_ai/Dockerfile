FROM python:3.11-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

# Every dependency ships a prebuilt wheel for amd64 and arm64, so no compiler
# or system packages are needed.
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY app ./app
COPY scripts ./scripts

# Same uid as the chart's podSecurityContext (chart/call-center-ai).
RUN useradd --uid 10001 --no-create-home --shell /usr/sbin/nologin app
USER 10001

EXPOSE 8000

CMD ["uvicorn", "app.adapters.inbound.api.main:app", "--host", "0.0.0.0", "--port", "8000"]
