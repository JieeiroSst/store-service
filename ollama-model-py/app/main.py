from contextlib import asynccontextmanager

from fastapi import FastAPI

from . import faq, routes
from .internal_api import intents as internal_intents


@asynccontextmanager
async def lifespan(app: FastAPI):
    # FAQ matching is a mandatory feature, not best-effort, so a failure to
    # embed the FAQ set here should fail startup rather than silently degrade.
    await faq.load()
    await internal_intents.load()
    yield


app = FastAPI(title="ollama-model-py", lifespan=lifespan)
app.include_router(routes.router)
