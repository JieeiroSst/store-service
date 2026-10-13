import axios, { AxiosInstance, isAxiosError } from 'axios'
import { HttpError } from './errors'

export interface StartonTransaction {
    id: string
    transactionHash?: string
    status?: string
    network?: string
    [key: string]: unknown
}

export interface DeployResult {
    smartContract: { id?: string; address: string; network?: string; [key: string]: unknown }
    transaction?: StartonTransaction
}

export interface CallOptions {
    network: string
    address: string
    functionName: string
    params: unknown[]
    signerWallet: string
    value?: string
    speed?: string
}

export class StartonClient {
    constructor(private readonly http: AxiosInstance) {}

    static create(baseURL: string, apiKey: string): StartonClient {
        return new StartonClient(axios.create({ baseURL, headers: { 'x-api-key': apiKey }, timeout: 60_000 }))
    }

    async uploadFile(file: Buffer, fileName: string, contentType = 'application/octet-stream', metadata?: Record<string, unknown>) {
        const form = new FormData()
        form.append('file', new Blob([file], { type: contentType }), fileName)
        if (metadata) form.append('metadata', JSON.stringify(metadata))
        return this.send<{ cid: string; [key: string]: unknown }>('post', '/ipfs/file', form)
    }

    async uploadJson(name: string, content: Record<string, unknown>) {
        return this.send<{ cid: string; [key: string]: unknown }>('post', '/ipfs/json', { name, content })
    }

    async deployFromTemplate(body: { templateId: string; name: string; description?: string; network: string; params: unknown[]; signerWallet: string; speed?: string }) {
        return this.send<DeployResult>('post', '/smart-contract/from-template', body)
    }

    async deployFromBytecode(body: { name: string; description?: string; network: string; abi: unknown[]; bytecode: string; params: unknown[]; signerWallet: string; speed?: string }) {
        return this.send<DeployResult>('post', '/smart-contract/from-bytecode', body)
    }

    async importExisting(body: { network: string; address: string; name: string; description?: string; abi: unknown[] }) {
        return this.send<Record<string, unknown>>('post', '/smart-contract/import-existing', body)
    }

    async call({ network, address, functionName, params, signerWallet, value, speed }: CallOptions) {
        return this.send<StartonTransaction>('post', `/smart-contract/${enc(network)}/${enc(address)}/call`, {
            functionName,
            params,
            signerWallet,
            speed: speed ?? 'average',
            ...(value ? { value } : {}),
        })
    }

    async read(network: string, address: string, functionName: string, params: unknown[]) {
        return this.send<{ response: unknown; [key: string]: unknown }>('post', `/smart-contract/${enc(network)}/${enc(address)}/read`, { functionName, params })
    }

    async getTransaction(id: string) {
        return this.send<StartonTransaction>('get', `/transaction/${enc(id)}`)
    }

    private async send<T>(method: 'get' | 'post', url: string, data?: unknown): Promise<T> {
        try {
            const res = await this.http.request<T>({ method, url, data })
            return res.data
        } catch (err) {
            if (isAxiosError(err)) {
                const status = err.response?.status
                const body = err.response?.data as { message?: unknown } | undefined
                const message = typeof body?.message === 'string' ? body.message : err.message
                throw new HttpError(status && status < 500 ? status : 502, `starton: ${message}`, body)
            }
            throw err
        }
    }
}

function enc(v: string): string {
    return encodeURIComponent(v)
}
