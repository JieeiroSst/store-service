"""HTTP calls to internal company services — the "tools" a recognized
question can trigger instead of going to the FAQ script or the LLM (see
./intents.py for the recognition side).

Each internal service gets its own small function here, e.g.
get_points_balance() below calls point-service's `GET /:id`. Point it at a
real instance via POINTS_SERVICE_BASE_URL and nothing else in this package
needs to change. Failures raise the same BackendError used by ../proxy/client.py
so routes.py's existing error handling covers this path too.
"""
from .. import config
from ..proxy.client import request_json

cfg = config.load()


async def get_points_balance(user_id: str) -> dict:
    """Calls point-service's `GET /:id` for the given user/account id.

    Returns the raw RewardPointDTO JSON, e.g.
    {"total_points": 120, "points_active": 100, "points_pending": 20, ...}.
    Raises BackendError (see ../proxy/client.py) if point-service is
    unreachable or returns an error — let that propagate up to routes.py
    rather than swallowing it, since a hallucinated point balance is worse
    than a clear error.
    """
    return await request_json(
        cfg.points_service_base_url,
        "GET",
        f"/{user_id}",
        timeout=10.0,
        error_label="points service",
    )
