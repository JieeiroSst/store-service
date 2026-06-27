"""The one place to wire in a real backend.

routes.py calls answer() for both /api/generate and /api/chat — replace the
body below with a real call (an LLM API, an internal service, a database
lookup, whatever should actually answer the question) and nothing else in
this project needs to change.
"""


async def answer(question: str) -> str:
    return (
        f"[ollama-model-py placeholder] You asked: {question!r} — "
        "wire up app/proxy/client.py:answer() to answer this for real."
    )
