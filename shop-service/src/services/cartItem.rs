use std::sync::Arc;

use async_trait::async_trait;
use crate::core::{CommonError, QueryParams, ResultPaging};

use crate::domain::{CartItem, UpdateCartItem, CartItemRepo};


#[async_trait]
pub trait CartItemService: Send + Sync { 
    async fn create(&self, cart: &CartItem) -> RepoResult<()>;
    async fn update(&self, cart_id: &u16, update_cartItem: &UpdateCartItem) -> RepoResult<CartItem>;
    async fn delete(&self, cart_id: &u16) -> RepoResult<()>;
}

pub struct CartItemServiceImpl {
    pub cart_item_repo: Arc<dyn CartItemRepo>,
}


#[async_trait]
impl CartItemService for CartItemServiceImpl {

}