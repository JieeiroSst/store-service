use async_trait::async_trait;
use uuid::Uuid;

use crate::domain::error::DomainError;

use super::model::{RewardCreateModel, RewardModel, RewardUpdateModel};

#[async_trait]
pub trait RewardRepository: Send + Sync {
    async fn find(
        &self,
        name: &Option<String>,
        page: &u32,
        page_size: &u32,
    ) -> Result<Option<(Vec<RewardModel>, u32)>, DomainError>;
    async fn find_by_id(&self, id: &Uuid) -> Result<Option<RewardModel>, DomainError>;
    async fn insert(
        &self,
        category_create_model: &RewardCreateModel,
    ) -> Result<RewardModel, DomainError>;
    async fn update_by_id(
        &self,
        id: &Uuid,
        category_update_model: &RewardUpdateModel,
    ) -> Result<RewardModel, DomainError>;
    async fn delete_by_id(&self, id: &Uuid) -> Result<(), DomainError>;
}