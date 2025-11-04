use thiserror::Error;

#[derive(Debug, Error)]
pub enum IndexerError {
    #[error("Invalid message: {0}")]
    InvalidMessage(String),

    #[error("Storage error: {0}")]
    StorageError(String),

    #[error("ParsingError: {0}")]
    ParsingError(String),

    #[error("ConnectionError: {0}")]
    ConnectionError(String),
}
