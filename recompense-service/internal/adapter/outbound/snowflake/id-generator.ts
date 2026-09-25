import { Worker } from 'snowflake-uuid';

import type { IdGenerator } from "../../../application/port/outbound/id-generator";

export class SnowflakeIdGenerator implements IdGenerator {
    private readonly worker: Worker;

    constructor(workerId: number = 0, datacenterId: number = 1) {
        this.worker = new Worker(workerId, datacenterId);
    }

    nextId(): string {
        return this.worker.nextId().toString();
    }
}
