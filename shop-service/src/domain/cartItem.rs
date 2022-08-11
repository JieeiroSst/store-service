use async_trait::async_trait;
use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

use crate::core::{QueryParams, RepoResult, ResultPaging};

#[derive(StructOfArray,Debug, Clone, Serialize, Deserialize)]
pub struct CartItem {
    pub id:  u16,
    pub cart_id: u16,
    pub product: Product,
    pub total: u16,
    pub count: u16,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Clone)]
pub struct UpdateCartItem {
    pub product: Product,
    pub total: u16,
    pub count: u16,
}

#[async_trait]
pub trait CartItemRepo: Send + Sync {
    async fn create(&self, cart: &CartItem) -> RepoResult<()>;
    async fn update(&self, cart_id: &u16, update_cartItem: &UpdateCartItem) -> RepoResult<CartItem>;
    async fn delete(&self, cart_id: &u16) -> RepoResult<()>;
}