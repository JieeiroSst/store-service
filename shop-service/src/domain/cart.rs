use async_trait::async_trait;
use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

use crate::core::{QueryParams, RepoResult, ResultPaging};

#[derive(StructOfArray,Debug, Clone, Serialize, Deserialize)]
pub struct Cart {
    pub id: u16,
    pub total: u16,
    pub user_id: u16,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
    pub destroy: bool
}

#[derive(Debug, Clone)]
pub struct UpdateCart {
    pub total: u16,
    pub user_id: u16,
}

#[async_trait]
pub trait CartRepo: Send + Sync {
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Cart>>;
    async fn find(&self, id: &u16) -> RepoResult<Cart>;
    async fn find_by_user(&self, user_id: &u16) -> RepoResult<Cart>;
    async fn update(&self, id: &u16, update_cart: &UserUpdate) -> RepoResult<Cart>;
    async fn delete(&self, user_id: &u16) -> RepoResult<()>;
    async fn create(&self, cart: &Cart) -> RepoResult<()>;
}