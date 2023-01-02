use std::sync::Arc;

use async_trait::async_trait;
use crate::core::{CommonError, QueryParams, ResultPaging};

use crate::domain::{Product, ProductRepo};

#[async_trait]
pub trait ProductService: Send + Sync {
    async fn Create(&self, product: Product) ->  RepoResult<()>;
    async fn Update(&self, id: u16, product: Product) ->  RepoResult<()>;
    async fn delete(&self, id: &u16) -> RepoResult<()>;
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Product>>;
    async fn find(&self, id: &u16) -> RepoResult<Product>;
}

pub struct ProductServiceImpl {
    pub product_repo: Arc<dyn ProductRepo>,
}

#[async_trait]
impl ProductService for ProductServiceImpl {
    async fn Create(&self, product: Product) ->  RepoResult<()> {}

    async fn Update(&self, id: u16, product: Product) ->  RepoResult<()> {}

    async fn delete(&self, id: &u16) -> RepoResult<()> {}

    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Product>> {}

    async fn find(&self, id: &u16) -> RepoResult<Product> {}
}