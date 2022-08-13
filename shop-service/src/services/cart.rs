use std::sync::Arc;

use async_trait::async_trait;
use crate::core::{CommonError, QueryParams, ResultPaging};

use crate::domain::{Cart, UpdateCart, CartRepo};

#[async_trait]
pub trait CartService: Send + Sync {
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Cart, CommonError>>;
    async fn find(&self, id: &u16) -> RepoResult<Cart, CommonError>;
    async fn find_by_user(&self, user_id: &u16) -> RepoResult<Cart, CommonError>;
    async fn update(&self, id: &u16, update_cart: &UserUpdate) -> RepoResult<Cart, CommonError>;
    async fn delete(&self, user_id: &u16) -> RepoResult<(), CommonError>;
    async fn create(&self, cart: &Cart) -> RepoResult<(), CommonError>;
}

pub struct CartServiceImpl {
    pub cart_repo: Arc<dyn CartRepo>,
}

#[async_trait]
impl CartService for CartServiceImpl {
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Cart, CommonError>> {
        self.cart_repo.get_all(params).await.map_err(|e| -> CommonError{e.into()})
    }

    async fn find(&self, id: &u16) -> RepoResult<Cart, CommonError> {
        self.cart_repo.find(id).await.map_err(|e| -> CommonError{e.into()})
    }

    async fn find_by_user(&self, user_id: &u16) -> RepoResult<Cart, CommonError> {
        self.cart_repo.find_by_user(user_id).await.map_err(|e| -> CommonError{e.into()})
    }

    async fn update(&self, id: &u16, update_cart: &UserUpdate) -> RepoResult<Cart, CommonError>{
        self.cart_repo.update(id, update_cart).await.map_err(|e| -> CommonError{e.into()})
    }

    async fn delete(&self, user_id: &u16) -> RepoResult<(), CommonError>{
        
    }

    async fn create(&self, cart: &Cart) -> RepoResult<(), CommonError> {
        let mut cloned = cart.clone();

        self.cart_repo.create(&cloned).await.map_err(|e| -> CommonError{e.into()})
    }
}