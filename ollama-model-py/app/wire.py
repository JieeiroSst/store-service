"""Pydantic models for Ollama's own wire format.

These must match what internal/proxy/ollama_client.go (the Go side) sends
and expects — see this repo's internal/ollamaapi/types.go for the
equivalent Go structs.
"""
from __future__ import annotations

from typing import Optional

from pydantic import BaseModel, Field


class ChatMessage(BaseModel):
    role: str
    content: str


class ChatRequest(BaseModel):
    model: str
    messages: list[ChatMessage] = Field(default_factory=list)
    stream: Optional[bool] = None


class ChatResponseChunk(BaseModel):
    model: str
    created_at: str
    message: ChatMessage
    done: bool
    done_reason: Optional[str] = None
    total_duration: Optional[int] = None
    load_duration: Optional[int] = None
    prompt_eval_count: Optional[int] = None
    prompt_eval_duration: Optional[int] = None
    eval_count: Optional[int] = None
    eval_duration: Optional[int] = None


class GenerateRequest(BaseModel):
    model: str
    prompt: str = ""
    stream: Optional[bool] = None


class GenerateResponseChunk(BaseModel):
    model: str
    created_at: str
    response: str
    done: bool
    done_reason: Optional[str] = None
    context: Optional[list[int]] = None
    total_duration: Optional[int] = None
    load_duration: Optional[int] = None
    prompt_eval_count: Optional[int] = None
    prompt_eval_duration: Optional[int] = None
    eval_count: Optional[int] = None
    eval_duration: Optional[int] = None


class ModelDetails(BaseModel):
    parent_model: str = ""
    format: str = "gguf"
    family: str = "ollama"
    families: list[str] = Field(default_factory=lambda: ["ollama"])
    parameter_size: str = "N/A"
    quantization_level: str = "N/A"


class TagsModel(BaseModel):
    name: str
    model: str
    modified_at: str
    size: int
    digest: str
    details: ModelDetails


class TagsResponse(BaseModel):
    models: list[TagsModel]


class ShowRequest(BaseModel):
    model: str


class ShowResponse(BaseModel):
    modelfile: str
    parameters: str
    template: str
    details: ModelDetails
    model_info: dict = Field(default_factory=dict)
    capabilities: list[str] = Field(default_factory=list)
