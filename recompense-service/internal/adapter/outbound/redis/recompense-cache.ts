import { createClient } from 'redis';

import type { Recompense } from "../../../domain/recompense";
import type { RecompenseCache } from "../../../application/port/outbound/recompense-cache";

export type RedisClient = ReturnType<typeof createClient>;

export function newRedis(url: string): RedisClient {
    const client = createClient({ url });
    client.on('error', (error) => console.error('redis:', error));
    return client;
}

const key = (id: string) => `recompense:${id}`;

export class RedisRecompenseCache implements RecompenseCache {
    constructor(private readonly client: RedisClient, private readonly ttlSeconds: number) { }

    async get(id: string): Promise<Recompense | undefined> {
        try {
            if (!this.client.isReady) return undefined;
            const hit = await this.client.get(key(id));
            return hit ? (JSON.parse(hit) as Recompense) : undefined;
        } catch (error) {
            console.error("cache get:", error);
            return undefined;
        }
    }

    async set(id: string, r: Recompense): Promise<void> {
        try {
            if (this.client.isReady) await this.client.set(key(id), JSON.stringify(r), { EX: this.ttlSeconds });
        } catch (error) {
            console.error("cache set:", error);
        }
    }

    async del(id: string): Promise<void> {
        try {
            if (this.client.isReady) await this.client.del(key(id));
        } catch (error) {
            console.error("cache del:", error);
        }
    }
}

export class NoopRecompenseCache implements RecompenseCache {
    async get(): Promise<undefined> { return undefined; }
    async set(): Promise<void> { }
    async del(): Promise<void> { }
}
