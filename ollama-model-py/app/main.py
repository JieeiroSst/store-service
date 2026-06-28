from contextlib import asynccontextmanager

from fastapi import FastAPI

from . import bank_context, faq, pdf_rag, routes
from .internal_api import intents as internal_intents


@asynccontextmanager
async def lifespan(app: FastAPI):
    bank_context.load()
    await faq.load()
    await internal_intents.load()
    pdf_rag.load()
    yield


app = FastAPI(title="ollama-model-py", lifespan=lifespan)
app.include_router(routes.router)
