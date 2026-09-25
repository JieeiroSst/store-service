import type { RecompenseService } from "../../../application/port/inbound/recompense";
import type { LoyaltyService } from "../../../application/port/inbound/loyalty";
import { NotEligibleError, ValidationError } from "../../../domain/recompense";
import type { RecompenseRequest } from "./dto";

function json(body: unknown, status = 200): Response {
    return Response.json(body, { status });
}

export class RecompenseHandler {
    constructor(
        private readonly service: RecompenseService,
        private readonly loyalty: LoyaltyService,
    ) { }

    async handle(request: Request): Promise<Response> {
        const url = new URL(request.url);
        const path = url.pathname.replace(/\/+$/, "") || "/";
        const method = request.method;

        try {
            if (path === "/healthz") return json({ status: "ok" });
            if (path === "/readyz") {
                await this.service.ready();
                return json({ status: "ready" });
            }

            if (path === "/api/recompenses") {
                if (method === "GET") {
                    return json(await this.service.list(
                        Number(url.searchParams.get("page") ?? 1),
                        Number(url.searchParams.get("limit") ?? 20),
                    ));
                }
                if (method === "POST") {
                    const dto = (await request.json()) as RecompenseRequest;
                    return json(await this.service.create(dto), 201);
                }
                return json({ error: "method not allowed" }, 405);
            }

            const points = path.match(/^\/api\/members\/(\d+)\/points$/);
            if (points && method === "GET") return json(await this.loyalty.status(Number(points[1])));

            const earn = path.match(/^\/api\/members\/(\d+)\/points\/earn$/);
            if (earn && method === "POST") {
                const body = (await request.json()) as { amount: number; idempotencyKey: string };
                return json(await this.loyalty.earn(Number(earn[1]), body.idempotencyKey, Number(body.amount)));
            }

            const best = path.match(/^\/api\/members\/(\d+)\/recompenses\/best$/);
            if (best && method === "GET") {
                const found = await this.loyalty.best(Number(best[1]), Number(url.searchParams.get("amount")));
                return found ? json(found) : json({ error: "no usable recompense" }, 404);
            }

            const redeem = path.match(/^\/api\/recompenses\/(\d+)\/redeem$/);
            if (redeem && method === "POST") {
                const body = (await request.json()) as { memberId: number; amount: number };
                const result = await this.loyalty.redeem(redeem[1]!, Number(body.memberId), Number(body.amount));
                return result ? json(result) : json({ error: "not found" }, 404);
            }

            const member = path.match(/^\/api\/recompenses\/member\/(\d+)$/);
            if (member && method === "GET") {
                return json(await this.service.listByMember(
                    Number(member[1]),
                    Number(url.searchParams.get("page") ?? 1),
                    Number(url.searchParams.get("limit") ?? 20),
                ));
            }

            const one = path.match(/^\/api\/recompenses\/(\d+)$/);
            if (one) {
                const id = one[1]!;
                if (method === "GET") {
                    const found = await this.service.get(id);
                    return found ? json(found) : json({ error: "not found" }, 404);
                }
                if (method === "PUT") {
                    const dto = (await request.json()) as RecompenseRequest;
                    const updated = await this.service.update(id, dto);
                    return updated ? json(updated) : json({ error: "not found" }, 404);
                }
                if (method === "DELETE") {
                    return (await this.service.delete(id)) ? new Response(null, { status: 204 }) : json({ error: "not found" }, 404);
                }
                return json({ error: "method not allowed" }, 405);
            }

            return json({ error: "not found" }, 404);
        } catch (error) {
            if (error instanceof ValidationError || error instanceof SyntaxError) {
                return json({ error: error.message }, 400);
            }
            if (error instanceof NotEligibleError) return json({ error: error.message }, 422);
            console.error("request failed:", error);
            return json({ error: "internal error" }, 500);
        }
    }
}
