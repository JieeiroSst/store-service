use std::{error::Error, sync::Arc};

use deadpool_postgres::Pool;
use futures::StreamExt;
use lapin::{
    options::{
        BasicAckOptions, BasicConsumeOptions, BasicNackOptions, BasicRejectOptions,
        ExchangeDeclareOptions, QueueBindOptions, QueueDeclareOptions,
    },
    types::{AMQPValue, FieldTable},
    Connection, ConnectionProperties,
};

use crate::{
    amqp::{config::get_config, dto::RewardMessage},
    domain::reward::{self, repository::RewardRepository},
    repository::reward::PgRewardRepository,
};

pub async fn run(pg_pool: Arc<Pool>) -> Result<(), Box<dyn Error>> {
    let config = get_config();

    let connection =
        Arc::new(Connection::connect(&config.amqp_addr, ConnectionProperties::default()).await?);

    connection.on_error(|err| {
        log::error!("{}", err);
        std::process::exit(1);
    });

    let declare_channel = connection.create_channel().await?;

    declare_channel
        .exchange_declare(
            "reward-dead-letter.exchange",
            lapin::ExchangeKind::Direct,
            ExchangeDeclareOptions::default(),
            FieldTable::default(),
        )
        .await?;
    declare_channel
        .queue_declare(
            "reward-dead-letter.queue",
            QueueDeclareOptions {
                durable: true,
                ..Default::default()
            },
            FieldTable::default(),
        )
        .await?;
    declare_channel
        .queue_bind(
            "reward-dead-letter.queue",
            "reward-dead-letter.exchange",
            "",
            QueueBindOptions::default(),
            FieldTable::default(),
        )
        .await?;

    declare_channel
        .exchange_declare(
            "reward.exchange",
            lapin::ExchangeKind::Direct,
            ExchangeDeclareOptions {
                durable: true,
                ..Default::default()
            },
            FieldTable::default(),
        )
        .await?;

    let mut queue_field = FieldTable::default();
    queue_field.insert(
        "x-dead-letter-exchange".into(),
        AMQPValue::LongString("reward-dead-letter.exchange".into()),
    );
    declare_channel
        .queue_declare(
            "reward.queue",
            QueueDeclareOptions {
                durable: true,
                ..Default::default()
            },
            queue_field,
        )
        .await?;
    declare_channel
        .queue_bind(
            "reward.queue",
            "reward.exchange",
            "",
            QueueBindOptions::default(),
            FieldTable::default(),
        )
        .await?;
    declare_channel.close(0, "declare channel fineshed").await?;

    let reward_repository: Arc<dyn RewardRepository> =
        Arc::new(PgRewardRepository::new(pg_pool.clone()));

    let consumer_channel = connection.create_channel().await?;
    let mut consumer = consumer_channel
        .basic_consume(
            "reward.queue",
            "consumer",
            BasicConsumeOptions::default(),
            FieldTable::default(),
        )
        .await?;

    log::info!("server listener reward.queue");
    while let Some(result) = consumer.next().await {
        if let Ok(delivery) = result {
            match serde_json::from_slice::<RewardMessage>(delivery.data.as_slice()) {
                Ok(reward_message) => {
                    match reward::resources::create::execute(
                        reward_repository.clone(),
                        reward_message.into(),
                    )
                    .await
                    {
                        Ok(_) => delivery.ack(BasicAckOptions::default()).await?,
                        Err(err) => {
                            log::error!("Nack {}", err);
                            delivery.nack(BasicNackOptions::default()).await?;
                        }
                    }
                }
                Err(err) => {
                    log::error!("Reject {}", err);
                    delivery.reject(BasicRejectOptions::default()).await?;
                }
            }
        }
    }

    Ok(())
}