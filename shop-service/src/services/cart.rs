use std::sync::Arc;

use async_trait::async_trait;
use crate::core::{CommonError, QueryParams, ResultPaging};

use crate::domain::{Cart, UpdateCart, CartRepo, DeleteCart};

#[async_trait]
pub trait CartService: Send + Sync {
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Cart, CommonError>>;
    async fn find(&self, id: &u16) -> RepoResult<Cart, CommonError>;
    async fn find_by_user(&self, user_id: &u16) -> RepoResult<Cart, CommonError>;
    async fn update(&self, id: &u16, update_cart: &UserUpdate) -> RepoResult<Cart, CommonError>;
    async fn delete(&self, user_id: &u16, delete_cart: &DeleteCart) -> RepoResult<(), CommonError>;
    async fn create(&self, cart: &Cart) -> RepoResult<(), CommonError>;
    async fn order(&self, destroy: bool) -> RepoResult<Cart, CommonError>;
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

    async fn delete(&self, user_id: &u16, delete_cart: &DeleteCart) -> RepoResult<(), CommonError>{
        self.cart_repo.delete(user_id, delete_cart).await.map_err(|e| -> CommonError{e.into()})
    }

    async fn create(&self, cart: &Cart) -> RepoResult<(), CommonError> {
        self.cart_repo.create(&cart).await.map_err(|e| -> CommonError{e.into()})
    }

    async fn order(&self, destroy: bool) -> RepoResult<Cart, CommonError>  {
        self.cart_repo.order(destroy).await.map_err(|e| -> CommonError{e.into()})
    }
}