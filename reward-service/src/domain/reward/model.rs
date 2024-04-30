use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct RewardModel {
    pub id: Uuid,
    pub name: String,
    pub description: String,
    pub points: i64,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[cfg(test)]
impl RewardModel {
    pub fn new(name: String, description: String, points: i64) -> Self {
        Self {
            id: Uuid::new_v4(),
            name: name,
            description: description,
            points: points,
            created_at: Utc::now(),
            updated_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct RewardCreateModel {
    pub id: Uuid,
    pub name: String,
    pub description: String,
    pub points: i64,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl RewardCreateModel {
    pub fn new(name: String, description: String, points: i64) -> Self {
        Self {
            id: Uuid::new_v4(),
            name: name,
            description: description,
            points: points,
            created_at: Utc::now(),
            updated_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct RewardUpdateModel {
    pub name: String,
    pub description: String,
    pub points: i64,
}

impl RewardUpdateModel {
    pub fn new(name: String, description: String, points: i64) -> Self {
        Self {
            name: name,
            description: description,
            points: points,
        }
    }
}
