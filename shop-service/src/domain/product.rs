use async_trait::async_trait;
use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

use crate::core::{QueryParams, RepoResult, ResultPaging};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Product {
    pub id: u16,
    pub name: String,
    pub description: String,
    pub price: u16,
    pub media_id: u16,
    pub destroy: bool,
    pub media: Media,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Clone)]
pub struct UpdateProduct {
    pub name: String,
    pub description: String,
    pub price: u16,
    pub media_id: u16,
    pub destroy: bool,
}

#[derive(Debug, Clone)]
pub struct DeleteProduct {
    pub destroy: bool
}

#[async_trait]
pub trait ProductRepo: Send + Sync {
    async fn Create(&self, product: Product) ->  RepoResult<()>;
    async fn Update(&self, id: u16, update_product: UpdateProduct) ->  RepoResult<()>;
    async fn delete(&self, id: &u1, delete_product: DeleteProduct) -> RepoResult<()>;
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Product>>;
    async fn find(&self, id: &u16) -> RepoResult<Product>;
}