from fastapi import FastAPI

from . import routes

app = FastAPI(title="ollama-model-py")
app.include_router(routes.router)
