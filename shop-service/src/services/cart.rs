use std::sync::Arc;

use async_trait::async_trait;
use crate::core::{CommonError, QueryParams, ResultPaging};

use crate::domain::{Cart, UpdateCart, CartRepo};

#[async_trait]
pub trait CartService: Send + Sync {
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Cart>>;
    async fn find(&self, id: &u16) -> RepoResult<Cart>;
    async fn find_by_user(&self, user_id: &u16) -> RepoResult<Cart>;
    async fn update(&self, id: &u16, update_cart: &UserUpdate) -> RepoResult<Cart>;
    async fn delete(&self, user_id: &u16) -> RepoResult<()>;
    async fn create(&self, cart: &Cart) -> RepoResult<()>;
}

pub struct CartServiceImpl {
    pub cart_repo: Arc<dyn CartRepo>,
}

#[async_trait]
impl CartService for CartServiceImpl {

}