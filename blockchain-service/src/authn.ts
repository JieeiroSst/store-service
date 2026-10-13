import crypto from 'crypto'
import { HttpError, forbidden, unauthenticated } from './errors'

export const INTERNAL_TOKEN_HEADER = 'x-internal-token'
const IDENTITY_TTL_MS = 30_000

export interface Principal {
    userId: string
    username?: string
    role?: string
    roles: string[]
    service: boolean
}

export function isAdmin(p: Principal): boolean {
    return p.service || [p.role, ...p.roles].some((r) => r === 'admin' || r === 'super_admin')
}

export function owns(p: Principal, userId: string): boolean {
    return isAdmin(p) || (p.userId !== '' && p.userId === userId)
}

export function requireAdmin(p: Principal): void {
    if (!isAdmin(p)) throw forbidden()
}

export function actorOf(p: Principal): string {
    return p.service ? 'service' : p.userId
}

type Fetch = typeof fetch

export class Authenticator {
    private readonly internal: Buffer[]
    private readonly cache = new Map<string, { p: Principal; expires: number }>()
    private warned = false

    constructor(private readonly userServiceUrl: string, internalTokens: string[], private readonly http: Fetch = fetch) {
        this.internal = internalTokens.map((t) => Buffer.from(t))
    }

    get enforcing(): boolean {
        return this.internal.length > 0
    }

    async authenticate(authorization?: string, internalToken?: string): Promise<Principal> {
        const internal = internalToken?.trim()
        if (internal) {
            if (this.validInternal(internal)) return { userId: '', roles: [], service: true }
            throw unauthenticated()
        }
        const auth = authorization?.trim() ?? ''
        if (auth.startsWith('Bearer ') && auth.slice(7).trim()) {
            return this.user(auth.slice(7).trim())
        }
        if (!this.enforcing) {
            if (!this.warned) {
                this.warned = true
                console.warn('authn: INTERNAL_API_TOKENS is not set, requests without credentials are accepted as internal calls')
            }
            return { userId: '', roles: [], service: true }
        }
        throw unauthenticated()
    }

    private validInternal(token: string): boolean {
        const got = Buffer.from(token)
        let match = false
        for (const want of this.internal) {
            if (want.length === got.length && crypto.timingSafeEqual(want, got)) match = true
        }
        return match
    }

    private async user(token: string): Promise<Principal> {
        const cached = this.cache.get(token)
        if (cached && cached.expires > Date.now()) return cached.p
        const userId = await this.validate(token)
        const claims = readClaims(token)
        const p: Principal = {
            userId,
            username: claims.username,
            role: claims.role,
            roles: Array.isArray(claims.roles) ? claims.roles : [],
            service: false,
        }
        if (this.cache.size > 10_000) this.cache.clear()
        this.cache.set(token, { p, expires: Date.now() + IDENTITY_TTL_MS })
        return p
    }

    private async validate(token: string): Promise<string> {
        let res: Response
        try {
            res = await this.http(`${this.userServiceUrl}/api/v1/validate`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ session_token: token }),
                signal: AbortSignal.timeout(3000),
            })
        } catch (err) {
            console.error('authn: validate session:', err)
            throw new HttpError(503, 'authentication service unavailable')
        }
        if (res.status === 401 || res.status === 403) throw unauthenticated()
        if (res.status !== 200) {
            console.error(`authn: validate session returned ${res.status}`)
            throw new HttpError(503, 'authentication service unavailable')
        }
        const out = (await res.json().catch(() => ({}))) as { valid?: boolean; user_id?: string; userId?: string }
        const id = out.user_id || out.userId || ''
        if (!out.valid || !id) throw unauthenticated()
        return id
    }
}

function readClaims(token: string): { username?: string; role?: string; roles?: string[] } {
    const parts = token.split('.')
    if (parts.length !== 3) return {}
    try {
        return JSON.parse(Buffer.from(parts[1], 'base64url').toString('utf8'))
    } catch {
        return {}
    }
}
