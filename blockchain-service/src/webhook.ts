import crypto from 'crypto'
import { TransactionUpdate, normalizeStatus } from './transactions'

export const SIGNATURE_HEADER = 'starton-signature'

export function sign(secret: string, body: Buffer | string): string {
    return crypto.createHmac('sha256', secret).update(body).digest('hex')
}

export function verifySignature(secret: string, body: Buffer, signature: string | undefined): boolean {
    if (!secret || !signature) return false
    const expected = Buffer.from(sign(secret, body), 'hex')
    let got: Buffer
    try {
        got = Buffer.from(signature.trim().replace(/^sha256=/, ''), 'hex')
    } catch {
        return false
    }
    return got.length === expected.length && crypto.timingSafeEqual(got, expected)
}

type Json = Record<string, unknown>

function obj(v: unknown): Json | undefined {
    return v && typeof v === 'object' && !Array.isArray(v) ? (v as Json) : undefined
}

function str(v: unknown): string | undefined {
    return typeof v === 'string' && v ? v : undefined
}

export function parseWebhook(body: unknown): TransactionUpdate | null {
    const root = obj(body)
    if (!root) return null
    const payload = obj(root.payload) ?? root
    const tx = obj(payload.transaction) ?? payload
    const event = str(root.event)
    const status = str(tx.status) ?? event
    const id = str(tx.id) ?? str(payload.transactionId)
    const transactionHash = str(tx.transactionHash) ?? str(payload.transactionHash)
    if (!status || (!id && !transactionHash)) return null
    const error = str(tx.error) ?? str(obj(tx.error)?.message) ?? str(tx.errorMessage)
    return {
        id,
        transactionHash,
        status: normalizeStatus(status),
        error,
        network: str(tx.network) ?? str(root.network),
        raw: { event, webhookId: str(root.id) },
    }
}
