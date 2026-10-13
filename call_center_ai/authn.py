import base64
import hmac
import json
import logging
import os
import threading
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from typing import Dict, Optional, Sequence, Tuple

INTERNAL_TOKEN_HEADER = "X-Internal-Token"
ADMIN_ROLES = {"admin", "super_admin"}
IDENTITY_TTL = 30.0

log = logging.getLogger("authn")


class AuthError(Exception):
    def __init__(self, status: int, message: str):
        super().__init__(message)
        self.status = status
        self.message = message


UNAUTHENTICATED = (401, "authentication required")
FORBIDDEN = (403, "forbidden")
UNAVAILABLE = (503, "authentication service unavailable")


@dataclass(frozen=True)
class Principal:
    user_id: str = ""
    username: str = ""
    role: str = ""
    roles: Tuple[str, ...] = ()
    service: bool = False

    def is_admin(self) -> bool:
        return self.service or self.role in ADMIN_ROLES or any(r in ADMIN_ROLES for r in self.roles)

    def owns(self, user_id) -> bool:
        return self.is_admin() or (self.user_id != "" and self.user_id == str(user_id))


def _claims(token: str) -> dict:
    parts = token.split(".")
    if len(parts) != 3:
        return {}
    payload = parts[1] + "=" * (-len(parts[1]) % 4)
    try:
        return json.loads(base64.urlsafe_b64decode(payload))
    except ValueError:
        return {}


class Authenticator:
    def __init__(self, user_service_url: str, internal_tokens: Sequence[str], timeout: float = 3.0):
        self.base_url = user_service_url.rstrip("/")
        self.internal = [t.strip().encode() for t in internal_tokens if t.strip()]
        self.timeout = timeout
        self._cache: Dict[str, Tuple[Principal, float]] = {}
        self._lock = threading.Lock()
        self._warned = False

    @classmethod
    def from_env(cls) -> "Authenticator":
        return cls(
            os.getenv("USER_SERVICE_URL", "http://user-api-svc"),
            os.getenv("INTERNAL_API_TOKENS", "").split(","),
        )

    @property
    def enforcing(self) -> bool:
        return bool(self.internal)

    def authenticate(self, authorization: Optional[str], internal_token: Optional[str]) -> Optional[Principal]:
        internal_token = (internal_token or "").strip()
        if internal_token:
            if any(hmac.compare_digest(internal_token.encode(), t) for t in self.internal):
                return Principal(service=True)
            raise AuthError(*UNAUTHENTICATED)
        authorization = (authorization or "").strip()
        if authorization.startswith("Bearer ") and authorization[7:].strip():
            return self._user(authorization[7:].strip())
        if not self.enforcing:
            if not self._warned:
                self._warned = True
                log.warning("INTERNAL_API_TOKENS is not set, requests without credentials are accepted as internal calls")
            return Principal(service=True)
        return None

    def _user(self, token: str) -> Principal:
        now = time.monotonic()
        with self._lock:
            hit = self._cache.get(token)
            if hit and hit[1] > now:
                return hit[0]
        user_id = self._validate(token)
        c = _claims(token)
        p = Principal(
            user_id=user_id,
            username=str(c.get("username", "")),
            role=str(c.get("role", "")),
            roles=tuple(str(r) for r in (c.get("roles") or [])),
        )
        with self._lock:
            if len(self._cache) > 10000:
                self._cache.clear()
            self._cache[token] = (p, now + IDENTITY_TTL)
        return p

    def _validate(self, token: str) -> str:
        req = urllib.request.Request(
            self.base_url + "/api/v1/validate",
            data=json.dumps({"session_token": token}).encode(),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                body = json.loads(resp.read() or b"{}")
        except urllib.error.HTTPError as e:
            if e.code in (401, 403):
                raise AuthError(*UNAUTHENTICATED)
            log.error("validate session returned %s", e.code)
            raise AuthError(*UNAVAILABLE)
        except (urllib.error.URLError, OSError, ValueError) as e:
            log.error("validate session: %s", e)
            raise AuthError(*UNAVAILABLE)
        user_id = str(body.get("user_id") or body.get("userId") or "")
        if not body.get("valid") or not user_id:
            raise AuthError(*UNAUTHENTICATED)
        return user_id


def require(principal: Optional[Principal]) -> Principal:
    if principal is None:
        raise AuthError(*UNAUTHENTICATED)
    return principal


def require_admin(principal: Optional[Principal]) -> Principal:
    p = require(principal)
    if not p.is_admin():
        raise AuthError(*FORBIDDEN)
    return p


def require_owner(principal: Optional[Principal], user_id) -> Principal:
    p = require(principal)
    if not p.owns(user_id):
        raise AuthError(*FORBIDDEN)
    return p


def internal_headers() -> Dict[str, str]:
    token = os.getenv("INTERNAL_API_TOKEN", "").strip()
    return {INTERNAL_TOKEN_HEADER: token} if token else {}
