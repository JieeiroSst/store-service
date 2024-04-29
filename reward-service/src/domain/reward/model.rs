use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct RewardModel {
    pub id: Uuid,
    pub name: String,
    pub description: String,
    pub points: uint,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

