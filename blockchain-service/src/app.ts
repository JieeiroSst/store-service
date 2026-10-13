import express, { NextFunction, Request, RequestHandler, Response } from 'express'
import multer from 'multer'
import { utils } from 'ethers'
import { z, ZodError } from 'zod'
import { Authenticator, INTERNAL_TOKEN_HEADER, Principal, actorOf, isAdmin, owns, requireAdmin } from './authn'
import { ArtifactRegistry, CONTRACT_NAMES } from './contracts'
import { HttpError, forbidden, notFound } from './errors'
import { StartonClient } from './starton'
import { TransactionService, normalizeStatus } from './transactions'
import { SIGNATURE_HEADER, parseWebhook, verifySignature } from './webhook'

export interface AppDeps {
    starton: StartonClient
    artifacts: ArtifactRegistry
    transactions: TransactionService
    authn: Authenticator
    defaultNetwork: string
    signerWallet: string
    webhookSecret: string
}

const address = z.string().regex(/^0x[0-9a-fA-F]{40}$/, 'invalid address')
const network = z.string().regex(/^[a-z0-9-]{2,64}$/, 'invalid network')
const params = z.array(z.unknown()).default([])
const speed = z.enum(['low', 'average', 'fast', 'fastest']).optional()
const weiValue = z.string().regex(/^\d+$/, 'value must be a wei integer string').optional()

const deploySchema = z.object({
    contract: z.enum(CONTRACT_NAMES),
    params,
    network: network.optional(),
    signerWallet: address.optional(),
    name: z.string().min(1).max(200).optional(),
    description: z.string().max(1000).optional(),
    speed,
})

const deployTemplateSchema = z.object({
    templateId: z.string().min(1),
    name: z.string().min(1).max(200),
    description: z.string().max(1000).optional(),
    params,
    network: network.optional(),
    signerWallet: address.optional(),
    speed,
})

const importSchema = z.object({
    contract: z.enum(CONTRACT_NAMES),
    address,
    network: network.optional(),
    name: z.string().min(1).max(200).optional(),
    description: z.string().max(1000).optional(),
})

const callSchema = z.object({
    functionName: z.string().min(1),
    params,
    value: weiValue,
    speed,
    signerWallet: address.optional(),
    metadata: z.record(z.unknown()).optional(),
})

const readSchema = z.object({
    functionName: z.string().min(1),
    params,
})

const nftMetadata = z.object({
    name: z.string().min(1).max(200),
    description: z.string().max(5000).default(''),
    image: z.string().min(1),
    attributes: z.array(z.record(z.unknown())).optional(),
})

const mintSchema = z.object({
    contractAddress: address,
    to: address,
    network: network.optional(),
    signerWallet: address.optional(),
    tokenUri: z.string().min(1).optional(),
    metadata: nftMetadata.optional(),
    speed,
}).refine((v) => Boolean(v.tokenUri) !== Boolean(v.metadata), { message: 'provide exactly one of tokenUri or metadata' })

const ipfsJsonSchema = z.object({
    name: z.string().min(1).max(200),
    content: z.record(z.unknown()),
})

const metadataFieldsSchema = z.object({
    name: z.string().min(1).max(200),
    description: z.string().max(5000).default(''),
    attributes: z.string().optional().transform((v, ctx) => {
        if (!v) return undefined
        try {
            const parsed = JSON.parse(v)
            if (Array.isArray(parsed)) return parsed as Record<string, unknown>[]
        } catch {}
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'attributes must be a JSON array' })
        return z.NEVER
    }),
})

const escrowTxSchema = z.object({
    signerWallet: address.optional(),
    speed,
})

const listQuery = z.object({
    status: z.string().optional(),
    contractAddress: address.optional(),
    requestedBy: z.string().optional(),
    limit: z.coerce.number().int().min(1).max(200).default(50),
    offset: z.coerce.number().int().min(0).default(0),
})

const ESCROW_STATUS = ['None', 'Funded', 'Released', 'Refunded']

type Handler = (req: Request, res: Response, principal: Principal) => Promise<unknown>

export function orderKey(orderId: string): string {
    return /^0x[0-9a-fA-F]{64}$/.test(orderId) ? orderId.toLowerCase() : utils.id(orderId)
}

function toJson(value: unknown): unknown {
    if (Array.isArray(value)) return value.map(toJson)
    if (value && typeof value === 'object') {
        const v = value as { type?: unknown; hex?: unknown }
        if (v.type === 'BigNumber' && typeof v.hex === 'string') return BigInt(v.hex).toString()
        return Object.fromEntries(Object.entries(value).map(([k, x]) => [k, toJson(x)]))
    }
    return value
}

export function createApp(deps: AppDeps) {
    const app = express()
    app.disable('x-powered-by')
    const upload = multer({ storage: multer.memoryStorage(), limits: { fileSize: 20 * 1024 * 1024, files: 1 } })

    const authed = (handler: Handler): RequestHandler => async (req, res, next) => {
        try {
            const principal = await deps.authn.authenticate(req.header('authorization'), req.header(INTERNAL_TOKEN_HEADER))
            const out = await handler(req, res, principal)
            if (!res.headersSent) res.json(out)
        } catch (err) {
            next(err)
        }
    }

    const admin = (handler: Handler): RequestHandler => authed(async (req, res, p) => {
        requireAdmin(p)
        return handler(req, res, p)
    })

    const signer = (wallet?: string): string => {
        const w = wallet || deps.signerWallet
        if (!w) throw new HttpError(400, 'signerWallet is required (or set WALLET_ADDRESS)')
        return w
    }

    const startonCall = async (p: Principal, opts: {
        kind: 'call' | 'mint'
        network: string
        address: string
        functionName: string
        params: unknown[]
        signerWallet?: string
        value?: string
        speed?: string
        metadata?: Record<string, unknown>
    }) => {
        const tx = await deps.starton.call({
            network: opts.network,
            address: opts.address,
            functionName: opts.functionName,
            params: opts.params,
            signerWallet: signer(opts.signerWallet),
            value: opts.value,
            speed: opts.speed,
        })
        return deps.transactions.track({
            kind: opts.kind,
            network: opts.network,
            contractAddress: opts.address,
            functionName: opts.functionName,
            requestedBy: actorOf(p),
            metadata: opts.metadata,
            transaction: tx,
        })
    }

    const target = (req: Request) => ({
        network: network.parse(req.params.network),
        address: address.parse(req.params.address),
    })

    app.get(['/health', '/api/v1/health'], (_req, res) => {
        res.json({ status: 'ok' })
    })

    app.post('/api/v1/webhooks/starton', express.raw({ type: '*/*', limit: '1mb' }), async (req, res, next) => {
        try {
            if (!deps.webhookSecret) throw new HttpError(503, 'webhook secret is not configured')
            const body = Buffer.isBuffer(req.body) ? req.body : Buffer.alloc(0)
            if (!verifySignature(deps.webhookSecret, body, req.header(SIGNATURE_HEADER))) throw new HttpError(401, 'invalid signature')
            let parsed: unknown
            try {
                parsed = JSON.parse(body.toString('utf8'))
            } catch {
                throw new HttpError(400, 'invalid JSON body')
            }
            const update = parseWebhook(parsed)
            if (!update) {
                res.status(202).json({ ignored: true })
                return
            }
            const tx = await deps.transactions.apply(update)
            res.json({ received: true, transaction: tx })
        } catch (err) {
            next(err)
        }
    })

    app.use('/api/v1', express.json({ limit: '1mb' }))

    app.post('/api/v1/ipfs/files', upload.single('file'), admin(async (req) => {
        if (!req.file) throw new HttpError(400, 'file is required (multipart field "file")')
        return deps.starton.uploadFile(req.file.buffer, req.file.originalname, req.file.mimetype)
    }))

    app.post('/api/v1/ipfs/json', admin(async (req) => {
        const body = ipfsJsonSchema.parse(req.body)
        return deps.starton.uploadJson(body.name, body.content)
    }))

    app.post('/api/v1/nfts/metadata', upload.single('file'), admin(async (req) => {
        if (!req.file) throw new HttpError(400, 'file is required (multipart field "file")')
        const fields = metadataFieldsSchema.parse(req.body)
        const image = await deps.starton.uploadFile(req.file.buffer, req.file.originalname, req.file.mimetype)
        const content = {
            name: fields.name,
            description: fields.description,
            image: `ipfs://${image.cid}`,
            ...(fields.attributes ? { attributes: fields.attributes } : {}),
        }
        const metadata = await deps.starton.uploadJson(fields.name, content)
        return { imageCid: image.cid, metadataCid: metadata.cid, tokenUri: metadata.cid, metadata: content }
    }))

    app.post('/api/v1/nfts/mint', admin(async (req, _res, p) => {
        const body = mintSchema.parse(req.body)
        let tokenUri = body.tokenUri
        if (body.metadata) {
            const uploaded = await deps.starton.uploadJson(body.metadata.name, body.metadata)
            tokenUri = uploaded.cid
        }
        return startonCall(p, {
            kind: 'mint',
            network: body.network ?? deps.defaultNetwork,
            address: body.contractAddress,
            functionName: 'mint(address,string)',
            params: [body.to, tokenUri],
            signerWallet: body.signerWallet,
            speed: body.speed,
            metadata: { to: body.to, tokenUri },
        })
    }))

    app.post('/api/v1/contracts/deploy', admin(async (req, _res, p) => {
        const body = deploySchema.parse(req.body)
        const artifact = deps.artifacts.get(body.contract)
        const net = body.network ?? deps.defaultNetwork
        const result = await deps.starton.deployFromBytecode({
            name: body.name ?? body.contract,
            description: body.description,
            network: net,
            abi: artifact.abi,
            bytecode: artifact.bytecode,
            params: body.params,
            signerWallet: signer(body.signerWallet),
            speed: body.speed,
        })
        const tx = result.transaction
            ? await deps.transactions.track({
                kind: 'deploy',
                network: net,
                contractAddress: result.smartContract.address,
                requestedBy: actorOf(p),
                metadata: { contract: body.contract, params: body.params },
                transaction: result.transaction,
            })
            : null
        return { smartContract: result.smartContract, transaction: tx }
    }))

    app.post('/api/v1/contracts/deploy-template', admin(async (req, _res, p) => {
        const body = deployTemplateSchema.parse(req.body)
        const net = body.network ?? deps.defaultNetwork
        const result = await deps.starton.deployFromTemplate({
            templateId: body.templateId,
            name: body.name,
            description: body.description,
            network: net,
            params: body.params,
            signerWallet: signer(body.signerWallet),
            speed: body.speed,
        })
        const tx = result.transaction
            ? await deps.transactions.track({
                kind: 'deploy',
                network: net,
                contractAddress: result.smartContract.address,
                requestedBy: actorOf(p),
                metadata: { templateId: body.templateId, params: body.params },
                transaction: result.transaction,
            })
            : null
        return { smartContract: result.smartContract, transaction: tx }
    }))

    app.post('/api/v1/contracts/import', admin(async (req) => {
        const body = importSchema.parse(req.body)
        return deps.starton.importExisting({
            network: body.network ?? deps.defaultNetwork,
            address: body.address,
            name: body.name ?? body.contract,
            description: body.description,
            abi: deps.artifacts.get(body.contract).abi,
        })
    }))

    app.post('/api/v1/contracts/:network/:address/call', admin(async (req, _res, p) => {
        const t = target(req)
        const body = callSchema.parse(req.body)
        return startonCall(p, { kind: 'call', ...t, ...body })
    }))

    app.post('/api/v1/contracts/:network/:address/read', authed(async (req) => {
        const t = target(req)
        const body = readSchema.parse(req.body)
        const out = await deps.starton.read(t.network, t.address, body.functionName, body.params)
        return toJson(out)
    }))

    app.get('/api/v1/marketplace/:network/:address/listings/:listingId', authed(async (req) => {
        const t = target(req)
        const id = z.coerce.number().int().min(1).parse(req.params.listingId)
        const out = await deps.starton.read(t.network, t.address, 'listings', [String(id)])
        return { listingId: id, listing: toJson(out.response) }
    }))

    app.get('/api/v1/escrow/:network/:address/orders/:orderId', authed(async (req) => {
        const t = target(req)
        const key = orderKey(req.params.orderId)
        const out = await deps.starton.read(t.network, t.address, 'payments', [key])
        const payment = toJson(out.response)
        const status = Array.isArray(payment) ? ESCROW_STATUS[Number(payment[5])] : undefined
        return { orderId: req.params.orderId, orderKey: key, payment, status }
    }))

    for (const action of ['release', 'refund'] as const) {
        app.post(`/api/v1/escrow/:network/:address/orders/:orderId/${action}`, admin(async (req, _res, p) => {
            const t = target(req)
            const body = escrowTxSchema.parse(req.body ?? {})
            const key = orderKey(req.params.orderId)
            return startonCall(p, {
                kind: 'call',
                ...t,
                functionName: action,
                params: [key],
                signerWallet: body.signerWallet,
                speed: body.speed,
                metadata: { orderId: req.params.orderId, orderKey: key, action },
            })
        }))
    }

    app.get('/api/v1/transactions', authed(async (req, _res, p) => {
        const q = listQuery.parse(req.query)
        const requestedBy = isAdmin(p) ? q.requestedBy : actorOf(p)
        return deps.transactions.list({
            requestedBy,
            status: q.status ? normalizeStatus(q.status) : undefined,
            contractAddress: q.contractAddress,
            limit: q.limit,
            offset: q.offset,
        })
    }))

    app.get('/api/v1/transactions/:id', authed(async (req, _res, p) => {
        const tx = await deps.transactions.get(req.params.id)
        if (!tx) throw notFound('transaction')
        if (!owns(p, tx.requestedBy)) throw forbidden()
        return tx
    }))

    app.post('/api/v1/transactions/:id/sync', admin(async (req) => {
        const existing = await deps.transactions.get(req.params.id)
        if (!existing) throw notFound('transaction')
        const remote = await deps.starton.getTransaction(existing.id)
        return deps.transactions.apply({
            id: existing.id,
            transactionHash: remote.transactionHash,
            status: normalizeStatus(remote.status),
        })
    }))

    app.use((_req, res) => {
        res.status(404).json({ error: 'not found' })
    })

    app.use((err: unknown, _req: Request, res: Response, _next: NextFunction) => {
        if (err instanceof ZodError) {
            res.status(400).json({ error: 'invalid request', details: err.issues })
            return
        }
        if (err instanceof HttpError) {
            res.status(err.status).json({ error: err.message, ...(err.details !== undefined ? { details: err.details } : {}) })
            return
        }
        if (err instanceof multer.MulterError) {
            res.status(400).json({ error: err.message })
            return
        }
        if ((err as { type?: string })?.type === 'entity.parse.failed') {
            res.status(400).json({ error: 'invalid JSON body' })
            return
        }
        console.error(err)
        res.status(500).json({ error: 'internal error' })
    })

    return app
}
