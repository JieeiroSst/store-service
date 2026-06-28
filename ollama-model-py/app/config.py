import os
from dataclasses import dataclass
from pathlib import Path

_DEFAULT_FAQ_FILE = str(Path(__file__).parent / "faq.json")
_DEFAULT_PDF_VECTOR_FILE = str(Path(__file__).parent.parent / "data" / "pdf_vectors.json")


@dataclass(frozen=True)
class Config:
    port: int
    model_name: str
    ollama_base_url: str
    ollama_target_model: str
    ollama_embed_model: str
    faq_file: str
    faq_similarity_threshold: float
    points_service_base_url: str
    internal_api_similarity_threshold: float
    pdf_vector_file: str
    pdf_similarity_threshold: float


def load() -> Config:
    return Config(
        port=int(os.environ.get("PORT", "11434")),
        model_name=os.environ.get("MODEL_NAME", "custom-model"),
        ollama_base_url=os.environ.get("OLLAMA_BASE_URL", "http://localhost:11434"),
        ollama_target_model=os.environ.get("OLLAMA_TARGET_MODEL", "bank-assistant"),
        ollama_embed_model=os.environ.get("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
        faq_file=os.environ.get("FAQ_FILE", _DEFAULT_FAQ_FILE),
        faq_similarity_threshold=float(os.environ.get("FAQ_SIMILARITY_THRESHOLD", "0.82")),
        points_service_base_url=os.environ.get(
            "POINTS_SERVICE_BASE_URL", "http://localhost:8089"
        ),
        internal_api_similarity_threshold=float(
            os.environ.get("INTERNAL_API_SIMILARITY_THRESHOLD", "0.85")
        ),
        pdf_vector_file=os.environ.get("PDF_VECTOR_FILE", _DEFAULT_PDF_VECTOR_FILE),
        pdf_similarity_threshold=float(os.environ.get("PDF_SIMILARITY_THRESHOLD", "0.65")),
    )
