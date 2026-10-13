from typing import Optional

from fastapi import HTTPException, Request

import authn

_authenticator: Optional[authn.Authenticator] = None


def authenticator() -> authn.Authenticator:
    global _authenticator
    if _authenticator is None:
        _authenticator = authn.Authenticator.from_env()
    return _authenticator


def _wrap(fn, *args):
    try:
        return fn(*args)
    except authn.AuthError as e:
        raise HTTPException(status_code=e.status, detail=e.message)


def optional_principal(request: Request) -> Optional[authn.Principal]:
    return _wrap(
        authenticator().authenticate,
        request.headers.get("Authorization"),
        request.headers.get(authn.INTERNAL_TOKEN_HEADER),
    )


def require_user(request: Request) -> authn.Principal:
    return _wrap(authn.require, optional_principal(request))


def require_admin(request: Request) -> authn.Principal:
    return _wrap(authn.require_admin, optional_principal(request))


def ensure_owner(principal: authn.Principal, user_id) -> None:
    _wrap(authn.require_owner, principal, user_id)
