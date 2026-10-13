import { expect } from 'chai'
import request from 'supertest'
import { createApp, orderKey } from './app'
import { Authenticator } from './authn'
import { ArtifactRegistry } from './contracts'
import { BlockchainEvent, EventPublisher } from './events'
import { CallOptions, StartonClient } from './starton'
import { MemoryTransactionStore, TransactionService } from './transactions'
import { sign } from './webhook'

const SECRET = 'whsec'
const WALLET = '0x1111111111111111111111111111111111111111'
const CONTRACT = '0x2222222222222222222222222222222222222222'
const USER_WALLET = '0x3333333333333333333333333333333333333333'

class FakeStarton {
    calls: CallOptions[] = []
    json: { name: string; content: Record<string, unknown> }[] = []
    deployed: Record<string, unknown>[] = []
    nextId = 1

    async uploadFile() {
        return { cid: 'image-cid' }
    }

    async uploadJson(name: string, content: Record<string, unknown>) {
        this.json.push({ name, content })
        return { cid: `json-cid-${this.json.length}` }
    }

    async call(opts: CallOptions) {
        this.calls.push(opts)
        return { id: `tx-${this.nextId++}`, status: 'INITIALIZED' }
    }

    async read(_network: string, _address: string, functionName: string) {
        if (functionName === 'payments') return { response: ['0xb', '0xs', '0x0', { type: 'BigNumber', hex: '0x0de0b6b3a7640000' }, 123, 1] }
        return { response: 'ok' }
    }

    async deployFromBytecode(body: Record<string, unknown>) {
        this.deployed.push(body)
        return { smartContract: { address: CONTRACT }, transaction: { id: 'deploy-tx', transactionHash: '0xabc', status: 'PENDING' } }
    }

    async getTransaction(id: string) {
        return { id, transactionHash: '0xsynced', status: 'SUCCESS' }
    }
}

class RecordingPublisher implements EventPublisher {
    events: BlockchainEvent[] = []
    async publish(e: BlockchainEvent) {
        this.events.push(e)
    }
    async close() {}
}

class FakeArtifacts {
    get(name: string) {
        return { contractName: name, abi: [{ type: 'constructor' }], bytecode: '0x6000' }
    }
}

function setup() {
    const starton = new FakeStarton()
    const events = new RecordingPublisher()
    const userFetch = (async (_url: unknown, init?: { body?: unknown }) => {
        const token = JSON.parse(String(init?.body)).session_token
        if (token === 'alice') return new Response(JSON.stringify({ valid: true, user_id: 'alice' }), { status: 200 })
        if (token === 'bob') return new Response(JSON.stringify({ valid: true, user_id: 'bob' }), { status: 200 })
        return new Response('{}', { status: 401 })
    }) as typeof fetch
    const app = createApp({
        starton: starton as unknown as StartonClient,
        artifacts: new FakeArtifacts() as unknown as ArtifactRegistry,
        transactions: new TransactionService(new MemoryTransactionStore(), events),
        authn: new Authenticator('http://users', ['svc-token'], userFetch),
        defaultNetwork: 'avalanche-fuji',
        signerWallet: WALLET,
        webhookSecret: SECRET,
    })
    return { app, starton, events }
}

const svc = { 'X-Internal-Token': 'svc-token' }

function webhook(app: ReturnType<typeof setup>['app'], body: unknown, secret = SECRET) {
    const raw = JSON.stringify(body)
    return request(app).post('/api/v1/webhooks/starton').set('Content-Type', 'application/json').set('starton-signature', sign(secret, raw)).send(raw)
}

describe('blockchain-service API', () => {
    it('serves health without auth', async () => {
        const { app } = setup()
        await request(app).get('/health').expect(200, { status: 'ok' })
    })

    it('rejects anonymous and wrong internal token', async () => {
        const { app } = setup()
        await request(app).get('/api/v1/transactions').expect(401)
        await request(app).get('/api/v1/transactions').set('X-Internal-Token', 'nope').expect(401)
        await request(app).get('/api/v1/transactions').set('Authorization', 'Bearer mallory').expect(401)
    })

    it('forbids users from spending operations', async () => {
        const { app, starton } = setup()
        await request(app).post('/api/v1/nfts/mint').set('Authorization', 'Bearer alice').send({ contractAddress: CONTRACT, to: USER_WALLET, tokenUri: 'cid' }).expect(403)
        expect(starton.calls).to.have.length(0)
    })

    it('mints with uploaded metadata and tracks the transaction', async () => {
        const { app, starton, events } = setup()
        const res = await request(app)
            .post('/api/v1/nfts/mint')
            .set(svc)
            .send({ contractAddress: CONTRACT, to: USER_WALLET, metadata: { name: 'Shirt', description: 'Limited', image: 'ipfs://image-cid' } })
            .expect(200)
        expect(starton.json[0].content.name).to.equal('Shirt')
        expect(starton.calls[0]).to.include({ network: 'avalanche-fuji', address: CONTRACT, functionName: 'mint(address,string)', signerWallet: WALLET })
        expect(starton.calls[0].params).to.deep.equal([USER_WALLET, 'json-cid-1'])
        expect(res.body).to.include({ id: 'tx-1', kind: 'mint', status: 'INITIALIZED', requestedBy: 'service' })
        expect(events.events.map((e) => e.type)).to.deep.equal(['transaction.created'])
    })

    it('validates mint input', async () => {
        const { app } = setup()
        const res = await request(app).post('/api/v1/nfts/mint').set(svc).send({ contractAddress: 'bad', to: USER_WALLET }).expect(400)
        expect(res.body.error).to.equal('invalid request')
        await request(app).post('/api/v1/nfts/mint').set(svc).send({ contractAddress: CONTRACT, to: USER_WALLET }).expect(400)
    })

    it('deploys a compiled contract through starton', async () => {
        const { app, starton } = setup()
        const res = await request(app).post('/api/v1/contracts/deploy').set(svc).send({ contract: 'PaymentEscrow', params: [WALLET] }).expect(200)
        expect(starton.deployed[0]).to.include({ name: 'PaymentEscrow', bytecode: '0x6000', network: 'avalanche-fuji' })
        expect(res.body.smartContract.address).to.equal(CONTRACT)
        expect(res.body.transaction).to.include({ id: 'deploy-tx', kind: 'deploy', status: 'PENDING' })
        await request(app).post('/api/v1/contracts/deploy').set(svc).send({ contract: 'Lock' }).expect(400)
    })

    it('arbitrates escrow orders by hashed order id', async () => {
        const { app, starton } = setup()
        await request(app).post(`/api/v1/escrow/avalanche-fuji/${CONTRACT}/orders/order-42/release`).set(svc).send({}).expect(200)
        expect(starton.calls[0]).to.include({ functionName: 'release' })
        expect(starton.calls[0].params).to.deep.equal([orderKey('order-42')])
        const read = await request(app).get(`/api/v1/escrow/avalanche-fuji/${CONTRACT}/orders/order-42`).set('Authorization', 'Bearer alice').expect(200)
        expect(read.body.status).to.equal('Funded')
        expect(read.body.payment[3]).to.equal('1000000000000000000')
    })

    it('applies signed webhooks and publishes updates', async () => {
        const { app, events } = setup()
        await request(app).post(`/api/v1/contracts/avalanche-fuji/${CONTRACT}/call`).set(svc).send({ functionName: 'pause' }).expect(200)
        await webhook(app, { event: 'TRANSACTION_SUCCESS', payload: { transaction: { id: 'tx-1', transactionHash: '0xhash', status: 'SUCCESS' } } }).expect(200)
        const tx = await request(app).get('/api/v1/transactions/tx-1').set(svc).expect(200)
        expect(tx.body).to.include({ status: 'SUCCESS', transactionHash: '0xhash' })
        await webhook(app, { event: 'TRANSACTION_PENDING', payload: { transaction: { id: 'tx-1', status: 'PENDING' } } }).expect(200)
        const after = await request(app).get('/api/v1/transactions/tx-1').set(svc).expect(200)
        expect(after.body.status).to.equal('SUCCESS')
        expect(events.events.map((e) => e.type)).to.deep.equal(['transaction.created', 'transaction.updated'])
    })

    it('rejects webhooks with bad signatures', async () => {
        const { app } = setup()
        await webhook(app, { event: 'TRANSACTION_SUCCESS', payload: { transaction: { id: 'x' } } }, 'wrong').expect(401)
        await request(app).post('/api/v1/webhooks/starton').send({}).expect(401)
    })

    it('records unknown transactions from webhooks as external', async () => {
        const { app } = setup()
        await webhook(app, { event: 'TRANSACTION_ERROR', network: 'polygon-amoy', payload: { transaction: { id: 'ext-1', error: { message: 'reverted' } } } }).expect(200)
        const tx = await request(app).get('/api/v1/transactions/ext-1').set(svc).expect(200)
        expect(tx.body).to.include({ kind: 'external', status: 'ERROR', network: 'polygon-amoy' })
    })

    it('scopes transaction reads to the owner', async () => {
        const { app } = setup()
        await webhook(app, { event: 'TRANSACTION_PENDING', payload: { transaction: { id: 'ext-2' } } }).expect(200)
        await request(app).get('/api/v1/transactions/ext-2').set('Authorization', 'Bearer alice').expect(403)
        const list = await request(app).get('/api/v1/transactions').set('Authorization', 'Bearer alice').expect(200)
        expect(list.body).to.deep.equal([])
        const all = await request(app).get('/api/v1/transactions').set(svc).expect(200)
        expect(all.body).to.have.length(1)
    })

    it('syncs a transaction from starton', async () => {
        const { app } = setup()
        await request(app).post(`/api/v1/contracts/avalanche-fuji/${CONTRACT}/call`).set(svc).send({ functionName: 'unpause' }).expect(200)
        const res = await request(app).post('/api/v1/transactions/tx-1/sync').set(svc).expect(200)
        expect(res.body).to.include({ status: 'SUCCESS', transactionHash: '0xsynced' })
        await request(app).post('/api/v1/transactions/nope/sync').set(svc).expect(404)
    })
})
