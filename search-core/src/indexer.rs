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
    text.split(|c: char| !c.is_alphanumeric())
        .filter(|s| !s.is_empty())
        .map(|s| s.to_string())
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::errors::IndexerError;
    use crate::models::{Message, Metadata};
    use mockall::{mock, predicate::*};
    use proptest::prelude::*;

    fn dummy_message() -> Message {
        Message {
            url: "https://example.com".into(),
            text: "Hello WORLD!".into(),
            metadata: Metadata {
                title: "Example".into(),
                description: "desc".into(),
                timestamp: "2025-11-05T00:00:00Z".into(),
                status_code: 200,
            },
        }
    }

    mock! {
        pub StorageMock {}
        impl crate::storage::Storage for StorageMock {
            fn store(&mut self, message: &Message) -> Result<(), IndexerError>;
        }
    }

    #[test]
    fn test_normalize_text() {
        let result = normalize_text("HeLLo WoRLd!");
        assert_eq!(result, "hello world!");
    }

    #[test]
    fn test_tokenize_basic() {
        let tokens = tokenize("Hello, world!!! this-is test123");
        assert_eq!(tokens, vec!["Hello", "world", "this", "is", "test123"]);
    }

    #[test]
    fn test_process_message_success() {
        let mut mock = MockStorageMock::new();
        mock.expect_store()
            .withf(|msg| msg.url == "https://example.com")
            .returning(|_| Ok(()));

        let msg = dummy_message();
        let res = process_message(&mut mock, &msg);
        assert!(res.is_ok());
    }

    #[test]
    fn test_process_message_invalid_status() {
        let mut mock = MockStorageMock::new();
        let mut msg = dummy_message();
        msg.metadata.status_code = 404;

        let res = process_message(&mut mock, &msg);
        assert!(res.is_err());
    }

    #[test]
    fn test_process_message_empty_text() {
        let mut mock = MockStorageMock::new();
        let mut msg = dummy_message();
        msg.text.clear();

        let res = process_message(&mut mock, &msg);
        assert!(res.is_err());
    }

    proptest! {
        #[test]
        fn prop_normalize_text_is_lowercase(input in "\\PC*") {
            let out = normalize_text(&input);
            prop_assert_eq!(out, input.to_lowercase());
        }

        #[test]
        fn prop_tokenize_removes_non_alphanumerics(input in "[A-Za-z0-9 ,.!?-]{0,30}") {
            let tokens = tokenize(&input);
            for t in tokens {
                prop_assert!(t.chars().all(|c| c.is_alphanumeric()));
            }
        }
    }
}
