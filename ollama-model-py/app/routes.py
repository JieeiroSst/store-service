import hashlib
import time
from datetime import datetime, timezone

from fastapi import APIRouter
from fastapi.responses import JSONResponse, StreamingResponse

from . import config
from .proxy.client import BackendError, answer, embed
from .wire import (
    ChatMessage,
    ChatRequest,
    ChatResponseChunk,
    EmbedRequest,
    EmbedResponse,
    GenerateRequest,
    GenerateResponseChunk,
    ModelDetails,
    ShowRequest,
    ShowResponse,
    TagsModel,
    TagsResponse,
)

router = APIRouter()
cfg = config.load()


def _now() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.%fZ")


def _digest(model_name: str) -> str:
    return hashlib.sha256(f"placeholder:{model_name}".encode()).hexdigest()


def _default_details() -> ModelDetails:
    return ModelDetails(format="gguf", family="ollama", families=["ollama"])


@router.post("/api/generate")
async def generate(req: GenerateRequest):
    try:
        text = await answer(req.prompt)
    except BackendError as exc:
        return JSONResponse(status_code=exc.status_code, content={"error": str(exc)})
    start = time.monotonic()

    if req.stream is None or req.stream:
        async def gen():
            chunk = GenerateResponseChunk(
                model=req.model, created_at=_now(), response=text, done=False
            )
            yield chunk.model_dump_json(exclude_none=True) + "\n"

            final = GenerateResponseChunk(
                model=req.model,
                created_at=_now(),
                response="",
                done=True,
                done_reason="stop",
                context=[],
                total_duration=int((time.monotonic() - start) * 1e9),
                prompt_eval_count=len(req.prompt.split()),
                eval_count=len(text.split()),
            )
            yield final.model_dump_json(exclude_none=True) + "\n"

        return StreamingResponse(gen(), media_type="application/x-ndjson")

    final = GenerateResponseChunk(
        model=req.model,
        created_at=_now(),
        response=text,
        done=True,
        done_reason="stop",
        context=[],
        total_duration=int((time.monotonic() - start) * 1e9),
        prompt_eval_count=len(req.prompt.split()),
        eval_count=len(text.split()),
    )
    return JSONResponse(content=final.model_dump(exclude_none=True))


@router.post("/api/chat")
async def chat(req: ChatRequest):
    question = req.messages[-1].content if req.messages else ""
    try:
        text = await answer(question)
    except BackendError as exc:
        return JSONResponse(status_code=exc.status_code, content={"error": str(exc)})
    start = time.monotonic()

    if req.stream is None or req.stream:
        async def gen():
            chunk = ChatResponseChunk(
                model=req.model,
                created_at=_now(),
                message=ChatMessage(role="assistant", content=text),
                done=False,
            )
            yield chunk.model_dump_json(exclude_none=True) + "\n"

            final = ChatResponseChunk(
                model=req.model,
                created_at=_now(),
                message=ChatMessage(role="assistant", content=""),
                done=True,
                done_reason="stop",
                total_duration=int((time.monotonic() - start) * 1e9),
                prompt_eval_count=len(question.split()),
                eval_count=len(text.split()),
            )
            yield final.model_dump_json(exclude_none=True) + "\n"

        return StreamingResponse(gen(), media_type="application/x-ndjson")

    final = ChatResponseChunk(
        model=req.model,
        created_at=_now(),
        message=ChatMessage(role="assistant", content=text),
        done=True,
        done_reason="stop",
        total_duration=int((time.monotonic() - start) * 1e9),
        prompt_eval_count=len(question.split()),
        eval_count=len(text.split()),
    )
    return JSONResponse(content=final.model_dump(exclude_none=True))


@router.post("/api/embed")
async def embeddings(req: EmbedRequest):
    inputs = [req.input] if isinstance(req.input, str) else req.input
    try:
        vectors = await embed(inputs)
    except BackendError as exc:
        return JSONResponse(status_code=exc.status_code, content={"error": str(exc)})

    return EmbedResponse(model=req.model, embeddings=vectors)


@router.get("/api/tags")
async def tags():
    return TagsResponse(
        models=[
            TagsModel(
                name=cfg.model_name,
                model=cfg.model_name,
                modified_at=_now(),
                size=0,
                digest=_digest(cfg.model_name),
                details=_default_details(),
            )
        ]
    )


@router.post("/api/show")
async def show(req: ShowRequest):
    if req.model != cfg.model_name:
        # Ollama's real error shape is a flat {"error": "..."}, not FastAPI's
        # default {"detail": "..."} — match it for client compatibility.
        return JSONResponse(
            status_code=404, content={"error": f"model '{req.model}' not found"}
        )

    return ShowResponse(
        modelfile=f"# Placeholder Python model (model: {cfg.model_name})",
        parameters="",
        template="{{ .Prompt }}",
        details=_default_details(),
        model_info={"general.architecture": "ollama"},
        capabilities=["completion"],
    )
