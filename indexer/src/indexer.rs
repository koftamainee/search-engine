use crate::errors::IndexerError;
use crate::models::Message;
use crate::storage::Storage;
use tracing::{error, info, warn};

const VALID_STATUS_CODES: [u16; 3] = [200, 201, 203];

pub fn process_message<S: Storage>(storage: &mut S, message: &Message) -> Result<(), IndexerError> {
    if message.url.is_empty() || message.text.is_empty() {
        warn!("Message rejected: URL or text is empty");
        return Err(IndexerError::InvalidMessage(
            "URL or text empty.".to_string(),
        ));
    }

    if !VALID_STATUS_CODES.contains(&message.metadata.status_code) {
        warn!(
            "Message rejected: invalid status code {}",
            message.metadata.status_code
        );
        return Err(IndexerError::InvalidMessage("Invalid status code".into()));
    }

    let normalized_text = normalize_text(&message.text);
    info!(
        "Text normalized: {}...",
        &normalized_text.chars().take(30).collect::<String>()
    );

    let msg = Message {
        text: normalized_text,
        ..message.clone()
    };

    match storage.store(&msg) {
        Ok(_) => info!("Message stored successfully"),
        Err(e) => {
            error!("Failed to store message: {}", e);
            return Err(e);
        }
    }

    Ok(())
}

pub fn normalize_text(text: &str) -> String {
    text.to_lowercase()
}

pub fn tokenize(text: &str) -> Vec<String> {
    text.split_whitespace()
        .map(|s| s.trim_matches(|c: char| !c.is_alphanumeric()))
        .filter(|s| !s.is_empty())
        .map(|s| s.to_string())
        .collect()
}
