use std::sync::Arc;

use uuid::Uuid;

use crate::domain::{
    reward::{
        model::{RewardModel, RewardUpdateModel},
        repository::RewardRepository,
    },
    error::DomainError,
};

pub async fn execute(
    reward_repository: Arc<dyn RewardRepository>,
    id: Uuid,
    reward_update_model: RewardUpdateModel,
) -> Result<RewardModel, DomainError> {
    let has_reward = reward_repository.find_by_id(&id).await?;
    if has_reward.is_none() {
        return Err(DomainError::NotFound(String::from("Category id not found")));
    }

    let reward = reward_repository
        .update_by_id(&id, &reward_update_model)
        .await?;

    Ok(reward)
}
