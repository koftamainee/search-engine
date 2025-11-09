use anyhow::{Context, Result};
use lapin::{Connection, ConnectionProperties};
use std::env;
use tracing::{error, info};

use lapin::{options::QueueDeclareOptions, types::FieldTable};
use search_core::consumer::consume_queue;
use search_core::storage::InMemoryStorage;

#[tokio::main]
async fn main() -> Result<()> {
    tracing_subscriber::fmt()
        .with_target(false)
        .compact()
        .init();

    for (key, value) in env::vars() {
        info!("{} = {}", key, value);
    }

    info!("Indexer starting up...");

    let amqp_url = env::var("RABBITMQ_URL").context("Missing RABBITMQ_URL")?;
    let queue_name = env::var("RABBITMQ_QUEUE").unwrap_or_else(|_| "crawler_queue".to_string());

    info!("Connecting to RabbitMQ at {}", amqp_url);

    let connection = Connection::connect(&amqp_url, ConnectionProperties::default())
        .await
        .context("Failed to connect to RabbitMQ")?;

    let channel = connection
        .create_channel()
        .await
        .context("Failed to create RabbitMQ channel")?;

    let queue = channel
        .queue_declare(
            &queue_name,
            QueueDeclareOptions {
                durable: true,
                exclusive: false,
                auto_delete: false,
                nowait: false,
                ..Default::default()
            },
            FieldTable::default(),
        )
        .await
        .context("Failed to declare queue")?;

    info!("Queue declared: {:?}", queue.name());

    //TODO: this is a placeholder, change later
    let mut storage = InMemoryStorage::default();

    info!(
        "Connected to RabbitMQ, starting to consume queue: {}",
        queue_name
    );

    if let Err(err) = consume_queue(&channel, &queue_name, &mut storage).await {
        error!("Fatal error while consuming queue: {:?}", err);
    }

    Ok(())
}
