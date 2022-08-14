use std::sync::Arc;

use async_trait::async_trait;
use chrono::NaiveDateTime;
use diesel::prelude::*;

use super::async_pool;
use super::errors::DieselRepoError;
use super::infra;
use super::schema::*;

use crate::core::{QueryParams, RepoResult, ResultPaging};
use crate::domain::product::{Product, UpdateProduct, DeleteProduct, ProductRepo};

#[derive(Queryable, Insertable)]
#[table_name = "products"]
pub struct ProductDiesel {
    pub id: u16,
    pub name: String,
    pub description: String,
    pub price: u16,
    pub media_id: u16,
    pub destroy: bool,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

impl Into<Product> for ProductDiesel {
    fn into(self) -> Product {
        Product {
            id: self.id,
            name: self.name,
            description: self.description,
            price: self.price,
            media_id: self.media_id,
            destroy: self.destroy,
            created_at: self.created_at,
            updated_at: self.updated_at,
        }
    }
}

impl From<Product> for ProductDiesel {
    fn from(p: Product) -> Self {
        ProductDiesel {
            id: self.id,
            name: self.name,
            description: self.description,
            price: self.price,
            media_id: self.media_id,
            destroy: self.destroy,
            created_at: self.created_at,
            updated_at: self.updated_at,
        }
    }
}

#[derive(Debug, Clone, AsChangeset)]
#[table_name = "products"]
pub struct UpdateProductDiesel {
    pub name: String,
    pub description: String,
    pub price: u16,
    pub media_id: u16,
    pub destroy: bool,
}

impl From<UpdateProduct> for UpdateProductDiesel {
    fn from(p: UpdateProduct) -> Self {
        UpdateProductDiesel {
            name: p.name,
            description: p.description,
            price: p.price,
            media_id: p.media_id,
            destroy: p.media_id,
        }
    }
}

#[derive(Debug, Clone, AsChangeset)]
#[table_name = "products"]
pub struct DeleteProductDiesel {
    pub destroy: bool
}

impl From<DeleteProduct> for DeleteProductDiesel {
    fn from(p: DeleteProduct) -> Self {
        DeleteProductDiesel {
            destroy: p.destroy,
        }
    }
}

pub struct ProductDieselImpl {
    pool: Arc<infra::DBConn>
}

impl ProductDieselImpl {
    pub fn new(db: Arc<infra::DBConn>) -> Self {
        ProductDieselImpl {
            pool: db,
        }
    }
}

impl ProductRepo for ProductDieselImpl {
    async fn Create(&self, product: Product) ->  RepoResult<()> {}

    async fn Update(&self, id: u16, update_product: UpdateProduct) ->  RepoResult<()> {}

    async fn delete(&self, id: &u1, delete_product: DeleteProduct) -> RepoResult<()> {}

    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Product>> {}

    async fn find(&self, id: &u16) -> RepoResult<Product> {}
}