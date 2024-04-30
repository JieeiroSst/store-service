use std::sync::Arc;

use crate::domain::{
    reward::{model::RewardModel, repository::RewardRepository},
    error::DomainError,
};

pub async fn execute(
    reward_repository: Arc<dyn RewardRepository>,
    name: Option<String>,
    page: u32,
    page_size: u32,
) -> Result<Option<(Vec<RewardModel>, u32)>, DomainError> {
    let rewards = reward_repository.find(&name, &page, &page_size).await?;

    if rewards.is_some() {
        return Ok(rewards);
    }

    Ok(None)
}
