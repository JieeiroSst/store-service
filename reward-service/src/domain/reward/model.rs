use chrono::{DateTime, Utc};
use uuid::Uuid;
extern crate time;

#[derive(Debug, Clone)]
pub struct RewardModel {
    pub id: Uuid,
    pub name: String,
    pub description: String,
    pub points: uint,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[cfg(test)]
impl RewardCreateModel {
    pub fn new(name: String, description: String, points: uint) -> Self {
        Self {
            id: Uuid::new_v4(),
            name: name,
            description: description,
            points: points,
            created_at: time::now(),
            updated_at: time::now(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct RewardCreateModel {
    pub id: Uuid,
    pub name: String,
    pub description: String,
    pub points: uint,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl RewardCreateModel {
    pub fn new(name: String, description: String, points: uint) -> Self {
        Self {
            id: Uuid::new_v4(),
            name: name,
            description: description,
            points: points,
            created_at: time::now(),
            updated_at: time::now(),
        }
    }
}

#[cfg(test)]
impl RewardCreateModel {
    pub fn mock_default() -> Self {
        Self {
            id: uuid::Uuid::new_v4(),
            name: "Burgers".to_string(),
            description: Some("The Big Burgers".to_string()),
            points: 10,
            created_at: time::now(),
            updated_at: time::now(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct RewardUpdateModel {
    pub name: String,
    pub description: String,
    pub points: uint,
}

impl RewardUpdateModel {
    pub fn new(name: String, description: String, points: uint) -> Self {
        Self {
            name: name,
            description: description,
            points: points,
        }
    }
}

#[cfg(test)]
impl RewardUpdateModel {
    pub fn mock_default() -> Self {
        Self {
            name: "Burgers".to_string(),
            description: Some("The Big Burgers".to_string()),
            points: 10,
        }
    }
}