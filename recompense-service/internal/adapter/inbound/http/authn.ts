export const INTERNAL_TOKEN_HEADER = "X-Internal-Token";
const ADMIN_ROLES = new Set(["admin", "super_admin"]);
const IDENTITY_TTL_MS = 30_000;

export class AuthError extends Error {
    constructor(readonly status: number, message: string) {
        super(message);
    }
}

export const unauthenticated = () => new AuthError(401, "authentication required");
export const forbidden = () => new AuthError(403, "forbidden");
const unavailable = () => new AuthError(503, "authentication service unavailable");

export class Principal {
    constructor(
        readonly userId = "",
        readonly role = "",
        readonly roles: string[] = [],
        readonly service = false,
    ) { }

    isAdmin(): boolean {
        return this.service || ADMIN_ROLES.has(this.role) || this.roles.some((r) => ADMIN_ROLES.has(r));
    }

    owns(userId: string | number): boolean {
        return this.isAdmin() || (this.userId !== "" && this.userId === String(userId));
    }
}

function claims(token: string): { role?: string; roles?: string[] } {
    const parts = token.split(".");
    if (parts.length !== 3) return {};
    try {
        return JSON.parse(Buffer.from(parts[1]!, "base64url").toString("utf8"));
    } catch {
        return {};
    }
}

function safeEqual(a: string, b: string): boolean {
    const x = Buffer.from(a);
    const y = Buffer.from(b);
    if (x.length !== y.length) return false;
    let diff = 0;
    for (let i = 0; i < x.length; i++) diff |= x[i]! ^ y[i]!;
    return diff === 0;
}

export class Authenticator {
    private readonly cache = new Map<string, { p: Principal; expires: number }>();
    private warned = false;

    constructor(
        private readonly userServiceURL: string,
        private readonly internalTokens: string[],
        private readonly timeoutMs = 3000,
    ) { }

    static fromEnv(): Authenticator {
        return new Authenticator(
            (process.env.USER_SERVICE_URL || "http://user-api-svc").replace(/\/+$/, ""),
            (process.env.INTERNAL_API_TOKENS || "").split(",").map((t) => t.trim()).filter(Boolean),
        );
    }

    get enforcing(): boolean {
        return this.internalTokens.length > 0;
    }

    async authenticate(request: Request): Promise<Principal | undefined> {
        const internal = (request.headers.get(INTERNAL_TOKEN_HEADER) || "").trim();
        if (internal) {
            if (this.internalTokens.some((t) => safeEqual(internal, t))) return new Principal("", "", [], true);
            throw unauthenticated();
        }
        const authorization = (request.headers.get("Authorization") || "").trim();
        if (authorization.startsWith("Bearer ") && authorization.slice(7).trim()) {
            return this.user(authorization.slice(7).trim());
        }
        if (!this.enforcing) {
            if (!this.warned) {
                this.warned = true;
                console.warn("authn: INTERNAL_API_TOKENS is not set, requests without credentials are accepted as internal calls");
            }
            return new Principal("", "", [], true);
        }
        return undefined;
    }

    private async user(token: string): Promise<Principal> {
        const hit = this.cache.get(token);
        if (hit && hit.expires > Date.now()) return hit.p;
        const userId = await this.validate(token);
        const c = claims(token);
        const p = new Principal(userId, String(c.role ?? ""), (c.roles ?? []).map(String));
        if (this.cache.size > 10_000) this.cache.clear();
        this.cache.set(token, { p, expires: Date.now() + IDENTITY_TTL_MS });
        return p;
    }

    private async validate(token: string): Promise<string> {
        let res: Response;
        try {
            res = await fetch(this.userServiceURL + "/api/v1/validate", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ session_token: token }),
                signal: AbortSignal.timeout(this.timeoutMs),
            });
        } catch (error) {
            console.error("authn: validate session:", error);
            throw unavailable();
        }
        if (res.status === 401 || res.status === 403) throw unauthenticated();
        if (!res.ok) {
            console.error("authn: validate session returned", res.status);
            throw unavailable();
        }
        const body = (await res.json().catch(() => ({}))) as { valid?: boolean; user_id?: string; userId?: string };
        const userId = String(body.user_id || body.userId || "");
        if (!body.valid || !userId) throw unauthenticated();
        return userId;
    }
}

export function requireUser(p: Principal | undefined): Principal {
    if (!p) throw unauthenticated();
    return p;
}

export function requireAdmin(p: Principal | undefined): Principal {
    const user = requireUser(p);
    if (!user.isAdmin()) throw forbidden();
    return user;
}

export function requireOwner(p: Principal | undefined, userId: string | number): Principal {
    const user = requireUser(p);
    if (!user.owns(userId)) throw forbidden();
    return user;
}
