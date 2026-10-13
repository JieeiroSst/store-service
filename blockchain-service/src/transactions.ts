import { Pool } from 'pg'
import { EventPublisher } from './events'
import { StartonTransaction } from './starton'

export type TransactionKind = 'deploy' | 'call' | 'mint' | 'external'

export interface TransactionRecord {
    id: string
    kind: TransactionKind
    network: string
    contractAddress: string | null
    functionName: string | null
    transactionHash: string | null
    status: string
    requestedBy: string
    metadata: Record<string, unknown>
    error: string | null
    createdAt: string
    updatedAt: string
}

export interface ListFilter {
    requestedBy?: string
    status?: string
    contractAddress?: string
    limit: number
    offset: number
}

export interface TransactionStore {
    insert(tx: TransactionRecord): Promise<void>
    get(id: string): Promise<TransactionRecord | null>
    findByHash(hash: string): Promise<TransactionRecord | null>
    update(id: string, patch: Partial<Pick<TransactionRecord, 'status' | 'transactionHash' | 'error'>>): Promise<TransactionRecord | null>
    list(filter: ListFilter): Promise<TransactionRecord[]>
}

const STATUS_RANK: Record<string, number> = {
    INITIALIZED: 0,
    PENDING: 1,
    REPLACED: 2,
    SUCCESS: 3,
    ERROR: 3,
    DROPPED: 3,
}

export function normalizeStatus(value: unknown): string {
    const s = String(value ?? '').trim().toUpperCase().replace(/^TRANSACTION_/, '')
    return s || 'PENDING'
}

function rank(status: string): number {
    return STATUS_RANK[status] ?? 1
}

export class MemoryTransactionStore implements TransactionStore {
    private readonly rows = new Map<string, TransactionRecord>()

    async insert(tx: TransactionRecord): Promise<void> {
        if (!this.rows.has(tx.id)) this.rows.set(tx.id, { ...tx })
    }

    async get(id: string): Promise<TransactionRecord | null> {
        const row = this.rows.get(id)
        return row ? { ...row } : null
    }

    async findByHash(hash: string): Promise<TransactionRecord | null> {
        const h = hash.toLowerCase()
        for (const row of this.rows.values()) {
            if (row.transactionHash?.toLowerCase() === h) return { ...row }
        }
        return null
    }

    async update(id: string, patch: Partial<Pick<TransactionRecord, 'status' | 'transactionHash' | 'error'>>): Promise<TransactionRecord | null> {
        const row = this.rows.get(id)
        if (!row) return null
        Object.assign(row, patch, { updatedAt: new Date().toISOString() })
        return { ...row }
    }

    async list(f: ListFilter): Promise<TransactionRecord[]> {
        return [...this.rows.values()]
            .filter((r) => !f.requestedBy || r.requestedBy === f.requestedBy)
            .filter((r) => !f.status || r.status === f.status)
            .filter((r) => !f.contractAddress || r.contractAddress?.toLowerCase() === f.contractAddress.toLowerCase())
            .sort((a, b) => b.createdAt.localeCompare(a.createdAt))
            .slice(f.offset, f.offset + f.limit)
            .map((r) => ({ ...r }))
    }
}

const COLUMNS = `id, kind, network, contract_address, function_name, transaction_hash, status, requested_by, metadata, error, created_at, updated_at`

interface Row {
    id: string
    kind: TransactionKind
    network: string
    contract_address: string | null
    function_name: string | null
    transaction_hash: string | null
    status: string
    requested_by: string
    metadata: Record<string, unknown>
    error: string | null
    created_at: Date
    updated_at: Date
}

function fromRow(r: Row): TransactionRecord {
    return {
        id: r.id,
        kind: r.kind,
        network: r.network,
        contractAddress: r.contract_address,
        functionName: r.function_name,
        transactionHash: r.transaction_hash,
        status: r.status,
        requestedBy: r.requested_by,
        metadata: r.metadata ?? {},
        error: r.error,
        createdAt: r.created_at.toISOString(),
        updatedAt: r.updated_at.toISOString(),
    }
}

export class PgTransactionStore implements TransactionStore {
    constructor(private readonly pool: Pool) {}

    async migrate(): Promise<void> {
        await this.pool.query(`
            CREATE TABLE IF NOT EXISTS blockchain_transactions (
                id TEXT PRIMARY KEY,
                kind TEXT NOT NULL,
                network TEXT NOT NULL,
                contract_address TEXT,
                function_name TEXT,
                transaction_hash TEXT,
                status TEXT NOT NULL,
                requested_by TEXT NOT NULL,
                metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
                error TEXT,
                created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
            );
            CREATE INDEX IF NOT EXISTS blockchain_transactions_hash_idx ON blockchain_transactions (lower(transaction_hash));
            CREATE INDEX IF NOT EXISTS blockchain_transactions_requested_by_idx ON blockchain_transactions (requested_by, created_at DESC);
        `)
    }

    async insert(tx: TransactionRecord): Promise<void> {
        await this.pool.query(
            `INSERT INTO blockchain_transactions (${COLUMNS})
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
             ON CONFLICT (id) DO NOTHING`,
            [tx.id, tx.kind, tx.network, tx.contractAddress, tx.functionName, tx.transactionHash, tx.status, tx.requestedBy, tx.metadata, tx.error, tx.createdAt, tx.updatedAt],
        )
    }

    async get(id: string): Promise<TransactionRecord | null> {
        const { rows } = await this.pool.query<Row>(`SELECT ${COLUMNS} FROM blockchain_transactions WHERE id = $1`, [id])
        return rows[0] ? fromRow(rows[0]) : null
    }

    async findByHash(hash: string): Promise<TransactionRecord | null> {
        const { rows } = await this.pool.query<Row>(
            `SELECT ${COLUMNS} FROM blockchain_transactions WHERE lower(transaction_hash) = lower($1) LIMIT 1`,
            [hash],
        )
        return rows[0] ? fromRow(rows[0]) : null
    }

    async update(id: string, patch: Partial<Pick<TransactionRecord, 'status' | 'transactionHash' | 'error'>>): Promise<TransactionRecord | null> {
        const { rows } = await this.pool.query<Row>(
            `UPDATE blockchain_transactions
             SET status = COALESCE($2, status),
                 transaction_hash = COALESCE($3, transaction_hash),
                 error = COALESCE($4, error),
                 updated_at = now()
             WHERE id = $1
             RETURNING ${COLUMNS}`,
            [id, patch.status ?? null, patch.transactionHash ?? null, patch.error ?? null],
        )
        return rows[0] ? fromRow(rows[0]) : null
    }

    async list(f: ListFilter): Promise<TransactionRecord[]> {
        const where: string[] = []
        const args: unknown[] = []
        if (f.requestedBy) {
            args.push(f.requestedBy)
            where.push(`requested_by = $${args.length}`)
        }
        if (f.status) {
            args.push(f.status)
            where.push(`status = $${args.length}`)
        }
        if (f.contractAddress) {
            args.push(f.contractAddress)
            where.push(`lower(contract_address) = lower($${args.length})`)
        }
        args.push(f.limit, f.offset)
        const { rows } = await this.pool.query<Row>(
            `SELECT ${COLUMNS} FROM blockchain_transactions
             ${where.length ? `WHERE ${where.join(' AND ')}` : ''}
             ORDER BY created_at DESC
             LIMIT $${args.length - 1} OFFSET $${args.length}`,
            args,
        )
        return rows.map(fromRow)
    }
}

export interface TrackInput {
    kind: TransactionKind
    network: string
    contractAddress?: string | null
    functionName?: string | null
    requestedBy: string
    metadata?: Record<string, unknown>
    transaction: StartonTransaction
}

export interface TransactionUpdate {
    id?: string
    transactionHash?: string
    status: string
    error?: string
    network?: string
    raw?: Record<string, unknown>
}

export class TransactionService {
    constructor(private readonly store: TransactionStore, private readonly events: EventPublisher) {}

    async track(input: TrackInput): Promise<TransactionRecord> {
        const now = new Date().toISOString()
        const record: TransactionRecord = {
            id: input.transaction.id,
            kind: input.kind,
            network: input.network,
            contractAddress: input.contractAddress ?? null,
            functionName: input.functionName ?? null,
            transactionHash: input.transaction.transactionHash ?? null,
            status: normalizeStatus(input.transaction.status),
            requestedBy: input.requestedBy,
            metadata: input.metadata ?? {},
            error: null,
            createdAt: now,
            updatedAt: now,
        }
        await this.store.insert(record)
        const stored = (await this.store.get(record.id)) ?? record
        await this.emit('transaction.created', stored)
        return stored
    }

    get(id: string): Promise<TransactionRecord | null> {
        return this.store.get(id)
    }

    list(filter: ListFilter): Promise<TransactionRecord[]> {
        return this.store.list(filter)
    }

    async apply(update: TransactionUpdate): Promise<TransactionRecord | null> {
        let current = update.id ? await this.store.get(update.id) : null
        if (!current && update.transactionHash) current = await this.store.findByHash(update.transactionHash)
        if (!current) {
            if (!update.id) return null
            return this.track({
                kind: 'external',
                network: update.network ?? 'unknown',
                requestedBy: 'external',
                metadata: update.raw ?? {},
                transaction: { id: update.id, transactionHash: update.transactionHash, status: update.status },
            })
        }
        const status = normalizeStatus(update.status)
        const regress = rank(current.status) === 3 ? status !== current.status : rank(status) < rank(current.status)
        const patch: Partial<Pick<TransactionRecord, 'status' | 'transactionHash' | 'error'>> = {}
        if (!regress && status !== current.status) patch.status = status
        if (update.transactionHash && update.transactionHash !== current.transactionHash) patch.transactionHash = update.transactionHash
        if (update.error && !regress) patch.error = update.error
        if (Object.keys(patch).length === 0) return current
        const updated = await this.store.update(current.id, patch)
        if (updated) await this.emit('transaction.updated', updated)
        return updated
    }

    private async emit(type: 'transaction.created' | 'transaction.updated', tx: TransactionRecord): Promise<void> {
        try {
            await this.events.publish({ type, key: tx.id, data: tx })
        } catch (err) {
            console.error(`events: publish ${type} ${tx.id} failed:`, err)
        }
    }
}
