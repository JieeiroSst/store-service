import logging
import os

from fastapi import HTTPException, Request
from twilio.request_validator import RequestValidator

log = logging.getLogger("twilio_auth")
_warned = False


def _public_url(request: Request) -> str:
    base = os.getenv("PUBLIC_BASE_URL", "").rstrip("/")
    if not base:
        return str(request.url)
    url = base + request.url.path
    if request.url.query:
        url += "?" + request.url.query
    return url


async def verify_twilio(request: Request) -> None:
    global _warned
    token = os.getenv("TWILIO_AUTH_TOKEN", "")
    if not token:
        if not _warned:
            _warned = True
            log.warning("TWILIO_AUTH_TOKEN is not set, Twilio webhook signatures are not verified")
        return
    form = await request.form()
    signature = request.headers.get("X-Twilio-Signature", "")
    if not RequestValidator(token).validate(_public_url(request), dict(form), signature):
        raise HTTPException(status_code=403, detail="invalid Twilio signature")
