use anyhow::{Context, Result};
use lapin::message::Delivery;
use lapin::options::BasicAckOptions;

use futures_util::stream::StreamExt;
use lapin::{Channel, options::BasicConsumeOptions, types::FieldTable};
use tracing::{error, info};

use serde_json::{Value, json};

use crate::errors::IndexerError;
use crate::indexer::process_message;
use crate::models::Message;
use crate::storage::Storage;

fn crawler_message_schema() -> Value {
    json!({
      "$schema": "http://json-schema.org/draft-07/schema#",
      "title": "CrawlerMessage",
      "type": "object",
      "properties": {
        "url": { "type": "string", "format": "uri" },
        "text": { "type": "string" },
        "metadata": {
          "type": "object",
          "properties": {
            "title": { "type": "string" },
            "description": { "type": "string" },
            "timestamp": { "type": "string", "format": "date-time" },
            "status_code": { "type": "integer" }
          },
          "required": ["title", "timestamp", "status_code"]
        }
      },
      "required": ["url", "text", "metadata"]
    })
}

pub async fn consume_queue<S: Storage>(
    channel: &Channel,
    queue_name: &str,
    storage: &mut S,
) -> Result<()> {
    let schema = crawler_message_schema();

    let mut consumer = channel
        .basic_consume(
            queue_name,
            "indexer_consumer",
            BasicConsumeOptions::default(),
            FieldTable::default(),
        )
        .await
        .context("Failed to start RabbitMQ consumer")?;

    info!("Started consuming queue '{}'", queue_name);

    while let Some(delivery_result) = consumer.next().await {
        match delivery_result {
            Ok(delivery) => {
                if let Err(err) = handle_delivery(&delivery, storage, &schema).await {
                    error!("Failed to process message: {:?}", err);
                    if let Err(e) = delivery.nack(Default::default()).await {
                        error!("Failed to nack message: {:?}", e);
                    }
                } else if let Err(e) = delivery.ack(BasicAckOptions::default()).await {
                    error!("Failed to ack message: {:?}", e);
                }
            }
            Err(err) => {
                error!("Failed to receive delivery: {:?}", err);
            }
        }
    }

    Ok(())
}

async fn handle_delivery<S: Storage>(
    delivery: &Delivery,
    storage: &mut S,
    schema: &Value,
) -> Result<()> {
    let data_str =
        std::str::from_utf8(&delivery.data).context("Failed to parse message as UTF-8")?;

    let instance: Value =
        serde_json::from_str(data_str).context("Failed to parse message as JSON")?;

    if !jsonschema::is_valid(schema, &instance) {
        return Err(IndexerError::InvalidMessage(
            "Message doesn't follow schema rules".to_string(),
        )
        .into());
    }

    let message: Message =
        serde_json::from_value(instance).context("Failed to deserialize into Message struct")?;

    process_message(storage, &message).context("Failed to process message")?;

    Ok(())
}
