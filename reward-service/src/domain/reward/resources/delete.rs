use std::sync::Arc;

use uuid::Uuid;

use crate::domain::{reward::repository::RewardRepository, error::DomainError};

pub async fn execute(
    reward_repository: Arc<dyn RewardRepository>,
    id: Uuid,
) -> Result<(), DomainError> {
    let has_category = reward_repository.find_by_id(&id).await?;
    if has_category.is_none() {
        return Err(DomainError::NotFound(String::from("reward id not found")));
    }

    reward_repository.delete_by_id(&category_id).await?;

    Ok(())
}