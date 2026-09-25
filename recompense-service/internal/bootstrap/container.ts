import type { Pool } from "pg";

import type { Config } from "../../config";
import { RecompenseServiceImpl } from "../application/service/recompense";
import type { RecompenseCache } from "../application/port/outbound/recompense-cache";
import { LoyaltyServiceImpl } from "../application/service/loyalty";
import { PostgresPointsLedger } from "../adapter/outbound/postgres/points-ledger";
import { RecompenseHandler } from "../adapter/inbound/http/handler";
import { migrate, newPool } from "../adapter/outbound/postgres/pool";
import { PostgresRecompenseRepository } from "../adapter/outbound/postgres/recompense-repository";
import { NoopRecompenseCache, RedisRecompenseCache, newRedis } from "../adapter/outbound/redis/recompense-cache";
import type { RedisClient } from "../adapter/outbound/redis/recompense-cache";
import { SnowflakeIdGenerator } from "../adapter/outbound/snowflake/id-generator";

export class Container {
    private constructor(
        readonly handler: RecompenseHandler,
        private readonly pool: Pool,
        private readonly redis: RedisClient | undefined,
    ) { }

    static async create(cfg: Config): Promise<Container> {
        const pool = newPool(cfg.Postgres);
        await migrate(cfg.Postgres, pool);

        let redis: RedisClient | undefined;
        let cache: RecompenseCache = new NoopRecompenseCache();
        if (cfg.RedisHost) {
            redis = newRedis(cfg.RedisHost);
            redis.connect().catch((error) => console.error("redis connect:", error));
            cache = new RedisRecompenseCache(redis, cfg.CacheTTLSeconds);
        }

        const repo = new PostgresRecompenseRepository(pool);
        const ids = new SnowflakeIdGenerator(Number(process.env.WORKER_ID ?? 0));
        const service = new RecompenseServiceImpl(repo, cache, ids);
        const loyalty = new LoyaltyServiceImpl(repo, new PostgresPointsLedger(pool), ids);
        return new Container(new RecompenseHandler(service, loyalty), pool, redis);
    }

    async close(): Promise<void> {
        if (this.redis?.isOpen) await this.redis.quit();
        await this.pool.end();
    }
}
