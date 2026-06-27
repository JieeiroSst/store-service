import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Config:
    port: int
    model_name: str
    ollama_base_url: str
    ollama_target_model: str


def load() -> Config:
    return Config(
        port=int(os.environ.get("PORT", "11434")),
        model_name=os.environ.get("MODEL_NAME", "custom-model"),
        ollama_base_url=os.environ.get("OLLAMA_BASE_URL", "http://localhost:11434"),
        ollama_target_model=os.environ.get("OLLAMA_TARGET_MODEL", "llama3.2:latest"),
    )
