use crate::errors::IndexerError;
use crate::models::Message;

pub trait Storage {
    fn store(&mut self, message: &Message) -> Result<(), IndexerError>;
}

#[derive(Default)]
pub struct InMemoryStorage {
    pub messages: Vec<Message>,
}

impl Storage for InMemoryStorage {
    fn store(&mut self, message: &Message) -> Result<(), IndexerError> {
        self.messages.push(message.clone());
        Ok(()) // NOTE: this is a placeholder
    }
}
