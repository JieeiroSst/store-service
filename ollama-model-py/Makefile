.PHONY: install run dev create-model ingest

OLLAMA_TARGET_MODEL ?= bank-assistant

install:
	python3 -m venv .venv
	.venv/bin/pip install -r requirements.txt

create-model:
	ollama create $(OLLAMA_TARGET_MODEL) -f Modelfile

ingest:
	. .venv/bin/activate && set -a && [ -f .env ] || cp .env.example .env && . .env && set +a && \
	python -m scripts.ingest_pdfs

run:
	. .venv/bin/activate && set -a && [ -f .env ] || cp .env.example .env && . .env && set +a && \
	uvicorn app.main:app --port "$${PORT:-11436}"

dev:
	. .venv/bin/activate && set -a && [ -f .env ] || cp .env.example .env && . .env && set +a && \
	uvicorn app.main:app --port "$${PORT:-11436}" --reload
