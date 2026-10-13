import path from 'path'

export interface Config {
    port: number
    startonBaseUrl: string
    startonApiKey: string
    defaultNetwork: string
    signerWallet: string
    webhookSecret: string
    databaseUrl: string
    kafkaBrokers: string[]
    kafkaTopic: string
    userServiceUrl: string
    internalTokens: string[]
    artifactsDir: string
}

function list(value: string | undefined): string[] {
    return (value ?? '').split(',').map((v) => v.trim()).filter(Boolean)
}

export function loadConfig(env: NodeJS.ProcessEnv = process.env): Config {
    return {
        port: Number(env.PORT || 8080),
        startonBaseUrl: env.STARTON_BASE_URL || 'https://api.starton.com/v3',
        startonApiKey: env.STARTON_API_KEY || '',
        defaultNetwork: env.STARTON_NETWORK || 'avalanche-fuji',
        signerWallet: env.WALLET_ADDRESS || '',
        webhookSecret: env.STARTON_WEBHOOK_SECRET || '',
        databaseUrl: env.DATABASE_URL || '',
        kafkaBrokers: list(env.KAFKA_BROKERS),
        kafkaTopic: env.KAFKA_TOPIC || 'blockchain.transactions',
        userServiceUrl: (env.USER_SERVICE_URL || 'http://user-api-svc').replace(/\/+$/, ''),
        internalTokens: list(env.INTERNAL_API_TOKENS),
        artifactsDir: env.ARTIFACTS_DIR || path.resolve(__dirname, '..', 'artifacts'),
    }
}
