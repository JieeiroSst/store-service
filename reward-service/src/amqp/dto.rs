use serde::Deserialize;

use crate::domain::reward::model::RewardCreateModel;

#[derive(Debug, Deserialize)]
pub struct RewardMessage {
    pub name: String,
    pub description: String,
    pub points: i64,
}

impl From<RewardMessage> for RewardCreateModel {
    fn from(value: RewardMessage) -> Self {
        RewardCreateModel::new(value.name, value.description, value.points)
    }
}