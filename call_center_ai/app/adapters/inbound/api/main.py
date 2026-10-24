from __future__ import annotations

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from app.adapters.inbound.api.routes import admin, voice
from app.config import settings
from app.container import Container
from app.domain.errors import AlreadyExistsError, NotFoundError

logging.basicConfig(
    level=settings.log_level,
    format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
)
log = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    container = getattr(app.state, "container", None) or Container.from_settings(settings)
    container.startup()
    app.state.container = container
    log.info("call center AI started")
    yield
    container.shutdown()


def create_app() -> FastAPI:
    app = FastAPI(
        title="Call Center AI API",
        description="Hệ thống call center với AI tự động trả lời",
        version="1.0.0",
        lifespan=lifespan,
    )
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=False,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    @app.exception_handler(NotFoundError)
    async def not_found(_: Request, exc: NotFoundError) -> JSONResponse:
        return JSONResponse(status_code=404, content={"detail": str(exc)})

    @app.exception_handler(AlreadyExistsError)
    async def already_exists(_: Request, exc: AlreadyExistsError) -> JSONResponse:
        return JSONResponse(status_code=400, content={"detail": str(exc)})

    @app.get("/")
    async def root() -> dict:
        return {"status": "healthy", "service": "Call Center AI", "version": "1.0.0"}

    app.include_router(voice.router)
    app.include_router(admin.router)
    return app


app = create_app()
