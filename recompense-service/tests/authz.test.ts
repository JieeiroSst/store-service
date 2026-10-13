import { afterAll, beforeAll, describe, expect, test } from "bun:test";

import { Authenticator } from "../internal/adapter/inbound/http/authn";
import { RecompenseHandler } from "../internal/adapter/inbound/http/handler";
import type { LoyaltyService } from "../internal/application/port/inbound/loyalty";
import type { RecompenseService } from "../internal/application/port/inbound/recompense";

const token = (sub: string, role: string) =>
    "h." + Buffer.from(JSON.stringify({ sub, role })).toString("base64url") + ".s";

let userService: ReturnType<typeof Bun.serve>;
let handler: RecompenseHandler;
const calls: string[] = [];

beforeAll(() => {
    userService = Bun.serve({
        port: 0,
        async fetch(req) {
            const { session_token } = (await req.json()) as { session_token: string };
            const parts = session_token.split(".");
            const sub = parts.length === 3 ? JSON.parse(Buffer.from(parts[1]!, "base64url").toString()).sub : "";
            return Response.json({ valid: Boolean(sub), user_id: sub });
        },
    });
    const service = {
        list: async () => { calls.push("list"); return []; },
        create: async () => { calls.push("create"); return {}; },
        listByMember: async () => { calls.push("listByMember"); return []; },
        get: async () => ({}),
        update: async () => ({}),
        delete: async () => true,
        ready: async () => undefined,
    } as unknown as RecompenseService;
    const loyalty = {
        status: async () => { calls.push("status"); return {}; },
        earn: async () => { calls.push("earn"); return {}; },
        best: async () => ({}),
        redeem: async () => { calls.push("redeem"); return {}; },
    } as unknown as LoyaltyService;
    handler = new RecompenseHandler(service, loyalty, new Authenticator(`http://localhost:${userService.port}`, ["s3cret"]));
});

afterAll(() => userService.stop());

async function status(method: string, path: string, opts: { bearer?: string; internal?: string; body?: unknown } = {}) {
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (opts.bearer) headers.Authorization = "Bearer " + opts.bearer;
    if (opts.internal) headers["X-Internal-Token"] = opts.internal;
    const res = await handler.handle(new Request("http://svc" + path, {
        method,
        headers,
        body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
    }));
    return res.status;
}

describe("authorization", () => {
    const user = token("7", "user");
    const admin = token("1", "admin");

    test("public catalogue", async () => {
        expect(await status("GET", "/api/recompenses")).toBe(200);
    });
    test("admin only writes", async () => {
        expect(await status("POST", "/api/recompenses", { body: {} })).toBe(401);
        expect(await status("POST", "/api/recompenses", { bearer: user, body: {} })).toBe(403);
        expect(await status("POST", "/api/recompenses", { bearer: admin, body: {} })).toBe(201);
        expect(await status("DELETE", "/api/recompenses/5", { bearer: user })).toBe(403);
    });
    test("member data is private", async () => {
        expect(await status("GET", "/api/members/7/points", { bearer: user })).toBe(200);
        expect(await status("GET", "/api/members/8/points", { bearer: user })).toBe(403);
        expect(await status("GET", "/api/recompenses/member/8", { bearer: user })).toBe(403);
        expect(await status("GET", "/api/members/8/recompenses/best?amount=10", { bearer: user })).toBe(403);
    });
    test("earning points is internal", async () => {
        expect(await status("POST", "/api/members/7/points/earn", { bearer: user, body: { amount: 1, idempotencyKey: "k" } })).toBe(403);
        expect(await status("POST", "/api/members/7/points/earn", { internal: "s3cret", body: { amount: 1, idempotencyKey: "k" } })).toBe(200);
        expect(await status("POST", "/api/members/7/points/earn", { internal: "wrong", body: {} })).toBe(401);
    });
    test("redeem only for own membership", async () => {
        expect(await status("POST", "/api/recompenses/5/redeem", { bearer: user, body: { memberId: 8, amount: 1 } })).toBe(403);
        expect(await status("POST", "/api/recompenses/5/redeem", { bearer: user, body: { memberId: 7, amount: 1 } })).toBe(200);
    });
    test("bad token rejected", async () => {
        expect(await status("GET", "/api/members/7/points", { bearer: "nope" })).toBe(401);
    });
});
