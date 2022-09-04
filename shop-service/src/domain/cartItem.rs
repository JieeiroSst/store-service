use async_trait::async_trait;
use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

use crate::core::{QueryParams, RepoResult, ResultPaging};
use crate::domain::Product;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CartItem {
    pub id:  u16,
    pub cart_id: u16,
    pub product: Product,
    pub total: u16,
    pub amount: u16,
    pub destroy: bool,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Clone)]
pub struct UpdateCartItem {
    pub total: u16,
    pub count: u16,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Clone)]
pub struct DeleteCartItem {
    pub destroy: bool
}

#[async_trait]
pub trait CartItemRepo: Send + Sync {
    async fn create(&self, cart: &CartItem) -> RepoResult<()>;
    async fn update(&self, cart_id: &u16, update_cartItem: &UpdateCartItem) -> RepoResult<CartItem>;
    async fn delete(&self, cart_id: &u16) -> RepoResult<()>;
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<CartItem>>;
    async fn find(&self, id: &u16) -> RepoResult<CartItem>;
}