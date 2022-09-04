use std::sync::Arc;

use async_trait::async_trait;
use chrono::NaiveDateTime;
use diesel::prelude::*;

use super::async_pool;
use super::error::DieselRepoError;
use super::infra;
use super::schema::*;

use crate::domain::Product;

use crate::core::{QueryParams, RepoResult, ResultPaging};
use crate::domain::cartItem::{CartItem, UpdateCartItem, DeleteCartItem, CartItemRepo};

#[derive(Queryable, Insertable)]
#[table_name= "cart_items"]
pub struct CartItemDiesel {
    pub id:  u16,
    pub cart_id: u16,
    pub total: u16,
    pub amount: u16,
    pub destroy: bool,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

impl Into<CartItem> for CartItemDiesel {
    fn into(self) -> CartItem {
        CartItem {
            id: self.id,
            cart_id: self.cart_id,
            total: self.total,
            amount: self.amount,
            destroy: self.destroy,
            created_at: self.created_at,
            updated_at: self.updated_at,
            product: todo!(),
        }
    }
}

impl From<CartItem> for CartItemDiesel {
    fn from(c: CartItem) -> Self {
        CartItemDiesel{
            id: c.id,
            cart_id: c.cart_id,
            total: c.total,
            amount: c.amount,
            destroy: c.destroy,
            created_at: c.created_at,
            updated_at: c.updated_at,
        }
    }
}

#[derive(Debug, Clone, AsChangeset)]
#[table_name = "cart_items"]
pub struct UpdateCartItemDiesel {
    pub total: u16,
    pub count: u16,
}

impl From<UpdateCartItem> for UpdateCartItemDiesel {
    fn from(u: UpdateCartItem) -> Self {
        UpdateCartItemDiesel{
            total: u.total,
            count: u.count,
        }
    }
}

#[derive(Debug, Clone, AsChangeset)]
#[table_name = "cart_items"]
pub struct DeleteCartItemDiesel {
    pub destroy: bool
}

impl From<DeleteCartItem> for DeleteCartItemDiesel {
    fn from(d: DeleteCartItem) -> Self {
        DeleteCartItemDiesel {
            destroy: d.destroy,
        }
    }
}

pub struct CartItemDieselImpl {
    pool: Arc<infra::DBConn>
}

impl CartItemDieselImpl {
    pub fn new(db: Arc<infra::DBConn>) -> Self {
        CartItemDieselImpl {
            pool: db,
        }
    }

    async fn total(&self) -> RepoResult<id64> {
        use super::schema::cart_items::dsl::cart_items;
        let pool = self.pool.clone();
        async_pool::run(move || {
            let conn = pool.get().unwrap;
            cart_items.count().get_result(&conn)
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())
        .map(|v: i64| v)
    }

    async fn fetch(&self, query: &dyn QueryParams) -> RepoResult<Vec<CartItem>> {
        use super::schema::cart_items::dsl::cart_items;
        let pool = self.pool.clone();
        let builder = cart_items.limit(query.limit()).offset(query.offset());
        let result = async_pool::run(move || {
            let conn = pool.get().unwrap();
            builder.load::<CartItemDiesel>(&conn)
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())?;
        OK(result.into_inner().map(|v| -> CartItem {v.into()}).collect())
    }
}

#[async_trait]
impl CartItemRepo for CartItemDieselImpl {
    async fn create(&self, cart: &CartItem) -> RepoResult<()> {
        let u = CartItemDiesel::from(cart.clone());
        use super::schema::cart_items::dsl::cart_items;

        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;
        async_pool::run(move || {
            diesel::insert_into(cart_items).value(u).execute(&conn)
            .await
            .map_err(|v| DieselRepoError::from(v).into_inner())?
        })
    }

    async fn update(&self, cart_id: &u16, update_cartItem: &UpdateCartItem) -> RepoResult<CartItem> {
        let u = CartItemDiesel::from(update_cartItem.clone());
        use super::schema::cart_items::dsl::{cart_id, cart_items};

        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;

        async_pool::run(move || {
            diesel::update(cart_items).filter(cart_id.eq(cart_id)).set(u).execute(&conn)
        }).await.map_err(|v| DieselRepoError::from(v).into_inner())?;
        self.find(cart_id).await
    }

    async fn delete(&self, cart_id: &u16, delete_cart_item: &DeleteCartItem) -> RepoResult<()> {
        let u = CartItemDieselImpl::from(delete_cart_item.clone());
        use super::schema::cart_items::dsl::{id, cart_items};
        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;

        async_pool::run(move || {
            diesel::update(cart_items).filter(cart_id.eq(cart_id)).set(u).execute(&conn)
        }).await.map_err(|v| DieselRepoError::from(v).into_inner())?
    }

    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<CartItem>> {
        let total = self.total();
        let carts = self.fetch(params);
        OK(ResultPaging {
            total: total.await?,
            items: carts.await?,
        })
    }

    async fn find(&self, id: &u16) -> RepoResult<CartItem> {
        use super::schema::cart_items::dsl::{id, cart_items};
        let conn = self.pool.get().map_err(|v| DieselRepoError::from::into_inner())?;

        async_pool::run(move || {
            OK(cart_items.filter(id.eq(id)).find::<CartItemDiesel>(&conn))
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())
        .map(|v| -> CartItem {v.into()})
    }
}