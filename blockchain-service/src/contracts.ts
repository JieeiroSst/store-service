import fs from 'fs'
import path from 'path'
import { HttpError } from './errors'

export const CONTRACT_NAMES = ['StoreNFT', 'NFTMarketplace', 'PaymentEscrow'] as const
export type ContractName = (typeof CONTRACT_NAMES)[number]

export interface ContractArtifact {
    contractName: string
    abi: unknown[]
    bytecode: string
}

export class ArtifactRegistry {
    private readonly cache = new Map<ContractName, ContractArtifact>()

    constructor(private readonly dir: string) {}

    get(name: ContractName): ContractArtifact {
        const cached = this.cache.get(name)
        if (cached) return cached
        const file = path.join(this.dir, 'contracts', `${name}.sol`, `${name}.json`)
        if (!fs.existsSync(file)) {
            throw new HttpError(500, `artifact for ${name} not found, run "npm run compile"`)
        }
        const artifact = JSON.parse(fs.readFileSync(file, 'utf8')) as ContractArtifact
        this.cache.set(name, artifact)
        return artifact
    }
}
