use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use utoipa::{IntoParams, ToSchema};
use uuid::Uuid;
use validator::Validate;

use crate::{
    api::utils::validator::validate_page_size_max,
    domain::reward::model::{RewardCreateModel, RewardModel, RewardUpdateModel},
};

#[cfg_attr(test, derive(Serialize))]
#[derive(Debug, Deserialize, Validate, ToSchema, Clone)]
pub struct RequestCreateReward {
    #[validate(length(max = 64))]
    pub name: String,
    #[validate(length(max = 512))]
    pub description: String,
    #[validate(length(max = 512))]
    pub points: uint,
}
impl From<RequestCreateReward> for RewardCreateModel {
    fn from(value: RequestCreateReward) -> Self {
        RewardCreateModel::new(value.name, value.description, value.points)
    }
}

#[cfg_attr(test, derive(Serialize))]
#[derive(Debug, Clone, Deserialize, Validate, ToSchema)]
pub struct RequestUpdateReward {
    #[validate(length(max = 64))]
    pub name: String,
    #[validate(length(max = 512))]
    pub description: String,
    #[validate(length(max = 512))]
    pub points: uint,
}
impl From<RequestUpdateReward> for RewardUpdateModel {
    fn from(value: RequestUpdateReward) -> Self {
        RewardUpdateModel::new(value.name, value.description, value.points)
    }
}

#[derive(Debug, Clone, Deserialize, Validate, IntoParams)]
pub struct RequestFindReward {
    #[validate(length(max = 64))]
    pub name: Option<String>,
    pub page: Option<u32>,
    #[validate(custom = "validate_page_size_max")]
    pub page_size: Option<u32>,
}

#[cfg_attr(test, derive(Deserialize))]
#[derive(Debug, Serialize, ToSchema)]
pub struct ResponseReward {
    pub id: Uuid,
    pub name: String,
    pub description: String,
    pub points: uint,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl From<RewardModel> for ResponseReward {
    fn from(value: RewardModel) -> Self {
        Self {
            id: value.id,
            name: value.name,
            description: value.description,
            points: value.points,
            created_at: value.created_at,
            updated_at: value.updated_at,
        }
    }
}