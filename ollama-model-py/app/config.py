import os
from dataclasses import dataclass
from pathlib import Path

_DEFAULT_FAQ_FILE = str(Path(__file__).parent / "faq.json")


@dataclass(frozen=True)
class Config:
    port: int
    model_name: str
    ollama_base_url: str
    ollama_target_model: str
    ollama_embed_model: str
    faq_file: str
    faq_similarity_threshold: float


def load() -> Config:
    return Config(
        port=int(os.environ.get("PORT", "11434")),
        model_name=os.environ.get("MODEL_NAME", "custom-model"),
        ollama_base_url=os.environ.get("OLLAMA_BASE_URL", "http://localhost:11434"),
        ollama_target_model=os.environ.get("OLLAMA_TARGET_MODEL", "llama3.2:latest"),
        ollama_embed_model=os.environ.get("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
        faq_file=os.environ.get("FAQ_FILE", _DEFAULT_FAQ_FILE),
        faq_similarity_threshold=float(os.environ.get("FAQ_SIMILARITY_THRESHOLD", "0.6")),
    )
