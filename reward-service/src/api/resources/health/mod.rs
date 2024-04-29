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
    pub points: uint,
}
impl From<RequestCreateReward> for CategoryCreateModel {
    fn from(value: RequestCreateReward) -> Self {
        CategoryCreateModel::new(value.name, value.description,value.points)
    }
}