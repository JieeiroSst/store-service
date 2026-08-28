"""FastAPI entrypoint (inbound adapter). Chạy bằng: uvicorn app.adapters.inbound.api.main:app"""
from __future__ import annotations

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.adapters.inbound.api.routes import router
from app.adapters.scheduler.daily_job import start_scheduler, stop_scheduler
from app.config import settings
from app.container import Container

logging.basicConfig(
    level=settings.log_level,
    format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Nếu app.state.container đã được set từ trước (vd test inject fake repository),
    # tôn trọng nó thay vì tạo kết nối Mongo thật — cho phép test FastAPI TestClient
    # chạy hoàn toàn offline.
    container: Container = getattr(app.state, "container", None) or Container(settings)
    if not container.has_repository:
        await container.startup()
    app.state.container = container

    scheduler = None
    if settings.scheduler_enabled:
        scheduler = start_scheduler(container, settings)
        logger.info("Scheduler đã bật, chạy hằng ngày lúc %s", settings.scheduler_run_hours)

    yield

    if scheduler is not None:
        stop_scheduler(scheduler)
    await container.disconnect()


app = FastAPI(
    title="Lottery Service",
    description="Cào & kiểm tra kết quả xổ số kiến thiết Việt Nam (3 miền)",
    version="1.0.0",
    lifespan=lifespan,
)
app.include_router(router)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok"}
