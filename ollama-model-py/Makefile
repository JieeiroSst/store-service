.PHONY: install run dev

install:
	python3 -m venv .venv
	.venv/bin/pip install -r requirements.txt

run:
	. .venv/bin/activate && set -a && [ -f .env ] || cp .env.example .env && . .env && set +a && \
	uvicorn app.main:app --port "$${PORT:-11434}"

dev:
	. .venv/bin/activate && set -a && [ -f .env ] || cp .env.example .env && . .env && set +a && \
	uvicorn app.main:app --port "$${PORT:-11434}" --reload
