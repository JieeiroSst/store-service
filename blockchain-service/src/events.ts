import { Kafka, Producer } from 'kafkajs'

export interface BlockchainEvent {
    type: 'transaction.created' | 'transaction.updated'
    key: string
    data: unknown
}

export interface EventPublisher {
    publish(event: BlockchainEvent): Promise<void>
    close(): Promise<void>
}

export class LogPublisher implements EventPublisher {
    async publish(event: BlockchainEvent): Promise<void> {
        console.log(`event ${event.type} ${event.key}`)
    }

    async close(): Promise<void> {}
}

export class KafkaPublisher implements EventPublisher {
    private connected?: Promise<void>

    private constructor(private readonly producer: Producer, private readonly topic: string) {}

    static create(brokers: string[], topic: string): KafkaPublisher {
        const kafka = new Kafka({ clientId: 'blockchain-service', brokers })
        return new KafkaPublisher(kafka.producer({ allowAutoTopicCreation: true }), topic)
    }

    async publish(event: BlockchainEvent): Promise<void> {
        this.connected ??= this.producer.connect().catch((err) => {
            this.connected = undefined
            throw err
        })
        await this.connected
        await this.producer.send({
            topic: this.topic,
            messages: [{
                key: event.key,
                value: JSON.stringify({ type: event.type, occurredAt: new Date().toISOString(), data: event.data }),
                headers: { 'event-type': event.type },
            }],
        })
    }

    async close(): Promise<void> {
        if (this.connected) await this.producer.disconnect()
    }
}
