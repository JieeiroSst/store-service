import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Config:
    port: int
    model_name: str


def load() -> Config:
    return Config(
        port=int(os.environ.get("PORT", "11434")),
        model_name=os.environ.get("MODEL_NAME", "custom-model"),
    )
