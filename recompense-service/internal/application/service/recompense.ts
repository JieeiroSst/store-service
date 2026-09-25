import type { RecompenseService } from "../port/inbound/recompense";
import type { RecompenseRepository } from "../port/outbound/recompense-repository";
import type { RecompenseCache } from "../port/outbound/recompense-cache";
import type { IdGenerator } from "../port/outbound/id-generator";
import { validate } from "../../domain/recompense";
import type { Recompense } from "../../domain/recompense";

const defaultLimit = 20;
const maxLimit = 100;

function clampLimit(limit: number): number {
    if (!Number.isFinite(limit) || limit <= 0) return defaultLimit;
    return Math.min(limit, maxLimit);
}

function offsetOf(page: number, limit: number): number {
    return (page > 0 ? page - 1 : 0) * limit;
}

export class RecompenseServiceImpl implements RecompenseService {
    constructor(
        private readonly repo: RecompenseRepository,
        private readonly cache: RecompenseCache,
        private readonly ids: IdGenerator,
    ) { }

    async create(input: Recompense): Promise<Recompense> {
        validate(input);
        return this.repo.insert({ ...input, ID: this.ids.nextId() });
    }

    async update(id: string, input: Recompense): Promise<Recompense | undefined> {
        validate(input);
        const updated = await this.repo.update(id, { ...input, ID: id });
        await this.cache.del(id);
        return updated;
    }

    async get(id: string): Promise<Recompense | undefined> {
        const hit = await this.cache.get(id);
        if (hit) return hit;
        const found = await this.repo.getByID(id);
        if (found) await this.cache.set(id, found);
        return found;
    }

    listByMember(memberId: number, page: number, limit: number): Promise<Recompense[]> {
        const l = clampLimit(limit);
        return this.repo.getByMemberID(memberId, l, offsetOf(page, l));
    }

    async delete(id: string): Promise<boolean> {
        const deleted = await this.repo.delete(id);
        await this.cache.del(id);
        return deleted;
    }

    list(page: number, limit: number): Promise<Recompense[]> {
        const l = clampLimit(limit);
        return this.repo.list(l, offsetOf(page, l));
    }

    ready(): Promise<void> {
        return this.repo.ping();
    }
}
