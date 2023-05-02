use std::sync::Arc;

use async_trait::async_trait;
use chrono::NaiveDateTime;
use diesel::prelude::*;
use tonic::codegen::ok;

use super::async_pool;
use super::error::DieselRepoError;
use super::infra;
use super::schema::*;

use crate::core::{QueryParams, RepoResult, ResultPaging};
use crate::domain::product::{Product, UpdateProduct, DeleteProduct, ProductRepo};
use crate::domain::Media;

#[derive(Queryable, Insertable)]
#[table_name = "products"]
pub struct ProductDiesel {
    pub id: String,
    pub product_name: String,
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
            media: todo!(),
        }
    }
}

impl From<Product> for ProductDiesel {
    fn from(p: Product) -> Self {
        ProductDiesel {
            id: p.id,
            product_name: p.product_name,
            description: p.description,
            price: p.price,
            media_id: p.media_id,
            destroy: p.destroy,
            created_at: p.created_at,
            updated_at: p.updated_at,
        }
    }
}

#[derive(Debug, Clone, AsChangeset)]
#[table_name = "products"]
pub struct UpdateProductDiesel {
    pub product_name: String,
    pub description: String,
    pub price: u16,
    pub media_id: String,
    pub destroy: bool,
    pub updated_at: NaiveDateTime,
}

impl From<UpdateProduct> for UpdateProductDiesel {
    fn from(p: UpdateProduct) -> Self {
        UpdateProductDiesel {
            product_name: p.product_name,
            description: p.description,
            price: p.price,
            media_id: p.media_id,
            destroy: p.media_id,
            updated_at: p.updated_at,
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

    async fn total(&self) -> RepoResult<i64> {
        use super::schema::products::dsl::products;
        let pool = self.pool.clone();
        async_pool::run(move || {
            let conn = pool.get().unwrap();
            products.count().get_result(&conn)
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())
        .map(|v: i64| v)
    }

    async fn fetch(&self, query: &dyn QueryParams) -> RepoResult<Vec<Product>> {
        use super::schema::products::dsl::products;
        let pool = self.pool.clone();
        let builder = products.limit(query.limit()).offset(query.offset());
        let result = async_pool::run(move || {
            let conn = pool.get().unwrap();
            builder.load::<ProductDiesel>(&conn)
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())?;
        Ok(result.into_iter().map(|v| -> Product { v.into() }).collect())
    }
}

#[async_trait]
impl ProductRepo for ProductDieselImpl {
    async fn create(&self, product: Product) ->  RepoResult<()> {
        let u = ProductDiesel::from(product.clone());
        use super::schema::products::dsl::products;

        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;
        async_pool::run(move || diesel::insert_into(products).values(u).execute(&conn))
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())?
    }

    async fn update(&self, id: u16, update_product: UpdateProduct) ->  RepoResult<()> {
        let u = ProductDiesel::from(update_product.clone());
        use super::schema::products::dsl::{id, products};

        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;

        async_pool::run(move || {
            diesel::update(products).filter(id.eq(id)).set(u).execute(&conn)
        }).await.map_err(|v| DieselRepoError::from(v).into_inner())?
    }

    async fn delete(&self, id: &u16, delete_product: DeleteProduct) -> RepoResult<()> {
        let u = DeleteProductDiesel::from(delete_product.clone());
        use super::schema::products::dsl::{id, products};
        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;

        async_pool::run(move || {
            diesel::update(products).filter(id.eq(id)).set(u).execute(&conn)
        }).await.map_err(|v| DieselRepoError::from(v).into_inner())?
    }

    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Product>> {
        let total = self.total();
        let carts = self.fetch(params);
        let result: ResultPaging = ResultPaging {
            total: total.await?,
            items: carts.await?,
        };
        Ok(result)
    }

    async fn find(&self, id: &u16) -> RepoResult<Product> {
        use super::schema::products::dsl::{id, products};
        let conn = self.pool.get().map_err(|v| DieselRepoError::from::into_inner())?;

        async_pool::run(move || {
            Ok(products.filter((id.eq(id)).first::<ProductDiesel>(&conn)))
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())
        .map(|v| -> Product {v.into()})
    }
}