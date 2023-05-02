use async_trait::async_trait;
use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

use crate::core::{QueryParams, RepoResult, ResultPaging};
use crate::domain::Media;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Product {
    pub id: String,
    pub name: String,
    pub description: String,
    pub price: u16,
    pub media_id: String,
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
    pub media_id: String,
    pub destroy: bool,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Clone)]
pub struct DeleteProduct {
    pub destroy: bool
}

#[async_trait]
pub trait ProductRepo: Send + Sync {
    async fn create(&self, product: Product) ->  RepoResult<()>;
    async fn update(&self, id: u16, update_product: UpdateProduct) ->  RepoResult<()>;
    async fn delete(&self, id: &u16, delete_product: DeleteProduct) -> RepoResult<()>;
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Product>>;
    async fn find(&self, id: &u16) -> RepoResult<Product>;
}