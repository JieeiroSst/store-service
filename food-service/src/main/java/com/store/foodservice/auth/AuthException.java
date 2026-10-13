package com.store.foodservice.auth;

public class AuthException extends RuntimeException {
    private final int status;

    public AuthException(int status, String message) {
        super(message);
        this.status = status;
    }

    public int status() {
        return status;
    }

    public static AuthException unauthenticated() {
        return new AuthException(401, "authentication required");
    }

    public static AuthException forbidden() {
        return new AuthException(403, "forbidden");
    }

    public static AuthException unavailable() {
        return new AuthException(503, "authentication service unavailable");
    }
}
