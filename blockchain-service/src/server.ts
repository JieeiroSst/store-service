import 'dotenv/config'
import { Pool } from 'pg'
import { createApp } from './app'
import { Authenticator } from './authn'
import { loadConfig } from './config'
import { ArtifactRegistry } from './contracts'
import { EventPublisher, KafkaPublisher, LogPublisher } from './events'
import { StartonClient } from './starton'
import { MemoryTransactionStore, PgTransactionStore, TransactionService, TransactionStore } from './transactions'

async function main() {
    const config = loadConfig()
    if (!config.startonApiKey) console.warn('STARTON_API_KEY is not set, Starton calls will fail')
    if (!config.webhookSecret) console.warn('STARTON_WEBHOOK_SECRET is not set, webhooks will be rejected')

    let pool: Pool | undefined
    let store: TransactionStore
    if (config.databaseUrl) {
        pool = new Pool({ connectionString: config.databaseUrl })
        const pg = new PgTransactionStore(pool)
        await pg.migrate()
        store = pg
    } else {
        console.warn('DATABASE_URL is not set, transactions are kept in memory')
        store = new MemoryTransactionStore()
    }

    const events: EventPublisher = config.kafkaBrokers.length
        ? KafkaPublisher.create(config.kafkaBrokers, config.kafkaTopic)
        : new LogPublisher()

    const app = createApp({
        starton: StartonClient.create(config.startonBaseUrl, config.startonApiKey),
        artifacts: new ArtifactRegistry(config.artifactsDir),
        transactions: new TransactionService(store, events),
        authn: new Authenticator(config.userServiceUrl, config.internalTokens),
        defaultNetwork: config.defaultNetwork,
        signerWallet: config.signerWallet,
        webhookSecret: config.webhookSecret,
    })

    const server = app.listen(config.port, () => {
        console.log(`blockchain-service listening on :${config.port}`)
    })

    const shutdown = (signal: string) => {
        console.log(`${signal} received, shutting down`)
        server.close(async () => {
            await events.close().catch((err) => console.error(err))
            await pool?.end().catch((err) => console.error(err))
            process.exit(0)
        })
        setTimeout(() => process.exit(1), 10_000).unref()
    }
    process.on('SIGTERM', () => shutdown('SIGTERM'))
    process.on('SIGINT', () => shutdown('SIGINT'))
}

main().catch((err) => {
    console.error(err)
    process.exit(1)
})
