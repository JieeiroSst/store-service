use std::sync::Arc;

use uuid::Uuid;

use crate::domain::{
    reward::{model::RewardModel, repository::RewardRepository},
    error::DomainError,
};

pub async fn execute(
    reward_repository: Arc<dyn RewardRepository>,
    id: Uuid,
) -> Result<Option<RewardModel>, DomainError> {
    if let Some(reward) = reward_repository.find_by_id(&id).await? {
        return Ok(Some(reward));
    }

    Ok(None)
}