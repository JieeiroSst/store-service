use std::sync::Arc;

use crate::domain::reward::model::RewardModel;
use crate::domain::{
    reward::{model::RewardCreateModel, repository::RewardRepository},
    error::DomainError,
};

pub async fn execute(
    reward_repository: Arc<dyn RewardRepository>,
    reward_create_model: RewardCreateModel,
) -> Result<RewardModel, DomainError> {
    let category = reward_repository.insert(&reward_create_model).await?;
    Ok(category)
}