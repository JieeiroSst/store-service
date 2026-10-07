FROM python:3.11-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY pyproject.toml poetry.lock README.md ./
COPY identified_service ./identified_service
RUN pip install --no-cache-dir .

RUN useradd -u 10001 -M app

USER app

EXPOSE 8000

CMD ["uvicorn", "identified_service.app:app", "--host", "0.0.0.0", "--port", "8000"]
