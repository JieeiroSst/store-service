"""Embed all PDFs under docs/ and save vectors to data/pdf_vectors.json.

Usage (from project root, after activating the venv):
    make ingest
    # or manually:
    set -a && . .env && set +a && python -m scripts.ingest_pdfs

Add/remove PDFs in docs/, run this again, then restart the service.

Folder layout expected:
    docs/
        banking-products/   ← product sheets, rate cards, brochures
        regulations/        ← NHNN circulars, compliance docs
        procedures/         ← internal guides, user manuals
"""
import asyncio
import json
import sys
from pathlib import Path

ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(ROOT))

from app.proxy.client import embed  # noqa: E402  (needs sys.path above)

DOCS_DIR = ROOT / "docs"
VECTOR_FILE = ROOT / "data" / "pdf_vectors.json"

CHUNK_SIZE = 600    # characters per chunk
CHUNK_OVERLAP = 120  # overlap between consecutive chunks
EMBED_BATCH = 32    # chunks per embedding request


def _extract_text(pdf_path: Path) -> str:
    from pypdf import PdfReader
    reader = PdfReader(str(pdf_path))
    pages = [page.extract_text() or "" for page in reader.pages]
    return "\n".join(pages)


def _chunk(text: str) -> list[str]:
    chunks: list[str] = []
    start = 0
    while start < len(text):
        piece = text[start : start + CHUNK_SIZE].strip()
        if piece:
            chunks.append(piece)
        start += CHUNK_SIZE - CHUNK_OVERLAP
    return chunks


async def main() -> None:
    pdf_files = sorted(DOCS_DIR.rglob("*.pdf"))
    if not pdf_files:
        print(f"No PDF files found under {DOCS_DIR}/")
        print("Drop your PDFs into one of the subfolders and run this again.")
        return

    records: list[dict] = []
    for pdf_path in pdf_files:
        rel = pdf_path.relative_to(ROOT)
        print(f"Reading {rel} …")
        text = _extract_text(pdf_path)
        chunks = _chunk(text)
        for chunk in chunks:
            records.append({"source": str(rel), "text": chunk})
        print(f"  {len(chunks)} chunks")

    texts = [r["text"] for r in records]
    print(f"\nEmbedding {len(texts)} chunks (batch={EMBED_BATCH}) …")

    embeddings: list[list[float]] = []
    for i in range(0, len(texts), EMBED_BATCH):
        batch = texts[i : i + EMBED_BATCH]
        batch_emb = await embed(batch)
        embeddings.extend(batch_emb)
        print(f"  {min(i + EMBED_BATCH, len(texts))}/{len(texts)}")

    for record, emb in zip(records, embeddings):
        record["embedding"] = emb

    VECTOR_FILE.parent.mkdir(exist_ok=True)
    with open(VECTOR_FILE, "w", encoding="utf-8") as f:
        json.dump(records, f, ensure_ascii=False)

    print(f"\nSaved {len(records)} vectors → {VECTOR_FILE.relative_to(ROOT)}")
    print("Restart the service to pick up the new vectors.")


if __name__ == "__main__":
    asyncio.run(main())
