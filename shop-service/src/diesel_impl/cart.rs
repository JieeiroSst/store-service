use std::sync::Arc;

use async_trait::async_trait;
use chrono::NaiveDateTime;
use diesel::prelude::*;

use super::async_pool;
use super::error::DieselRepoError;
use super::infra;
use super::schema::*;

use crate::core::{QueryParams, RepoResult, ResultPaging};
use crate::domain::cart::{Cart, UpdateCart, DeleteCart, CartRepo};

#[derive(Queryable, Insertable)]
#[table_name= "carts"]
pub struct CartDiesel {
    pub id: String,
    pub total: u16,
    pub user_id: String,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
    pub destroy: bool,
}

impl Into<Cart> for CartDiesel {
    fn into(self) -> Cart {
        Cart {
            id: self.id,
            total: self.total,
            user_id: self.user_id,
            created_at: self.created_at,
            updated_at: self.updated_at,
            destroy: self.destroy,
        }
    }
}

impl From<Cart> for CartDiesel {
    fn from(c: Cart) -> Self {
        CartDiesel {
            id: c.id,
            total: c.total,
            user_id: c.user_id,
            created_at: c.created_at,
            updated_at: c.updated_at,
            destroy: c.destroy,
        }
    }
}

#[derive(Debug, Clone, AsChangeset)]
#[table_name = "carts"]
pub struct CartUpdateDiesel {
    pub total: u16,
    pub user_id: String,
    pub destroy: bool,
    pub updated_at: NaiveDateTime,
}

impl From<UpdateCart> for CartUpdateDiesel {
    fn from(u :UpdateCart) -> Self {
        CartUpdateDiesel {
            total: u.total,
            user_id: u.user_id,
            destroy: u.destroy,
            updated_at: u.updated_at,
        }
    }
}

#[derive(Debug, Clone, AsChangeset)]
#[table_name = "carts"]
pub struct DeleteCartDiesel {
    pub destroy: bool
}

impl From<DeleteCart> for DeleteCartDiesel {
    fn from(d: DeleteCart) -> Self {
        DeleteCartDiesel {
            destroy: d.destroy,
        }
    }
}

pub struct CartDieselImpl {
    pool: Arc<infra::DBConn>
}

impl CartDieselImpl {
    pub fn new(db: Arc<infra::DBConn>) -> Self {
        CartDieselImpl{
            pool: db,
        }
    }

    async fn total(&self) -> RepoResult<i64> {
        use super::schema::carts::dsl::carts;
        let pool = self.pool.clone();
        async_pool::run(move || {
            let conn = pool.get().unwrap();
            carts.count().get_result(&conn)
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())
        .map(|v: i64| v)
    }

    async fn fetch(&self, query: &dyn QueryParams) -> RepoResult<Vec<Cart>> {
        use super::schema::carts::dsl::carts;
        let pool = self.pool.clone();
        let builder = carts.limit(query.limit()).offset(query.offset());
        let result = async_pool::run(move || {
            let conn = pool.get().unwrap();
            builder.load::<CartDiesel>(&conn)
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())?;
        Ok(result.into_iter().map(|v| -> Cart { v.into() }).collect())
    }
}

#[async_trait]
impl CartRepo for CartDieselImpl {
    async fn get_all(&self, params: &dyn QueryParams) -> RepoResult<ResultPaging<Cart>> {
        let total = self.total();
        let carts = self.fetch(params);

        let result: ResultPaging = ResultPaging {
            total: total.await?,
            items: carts.await?,
        };
        Ok(result)
    }

    async fn find(&self, id: &u16) -> RepoResult<Cart> {
        use super::schema::carts::dsl::{id, carts};
        let conn = self.pool.get().map_err(|v| DieselRepoError::from::into_inner())?;

        async_pool::run(move || {
            carts.filter(id.eq(id)).first::<CartDiesel>(&conn)
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())
        .map(|v| -> Cart {v.into()})
    }

    async fn find_by_user(&self, user_id: &u16) -> RepoResult<Cart> {
        use super::schema::carts::dsl::{user_id, carts};

        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;
        async_pool::run(move || {
            carts.filter(user_id.eq(user_id)).first::<CartDiesel>(&conn)
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())
        .map(|v| -> Cart {v.into()})
    }

    async fn update(&self, id: &u16, update_cart: &UpdateCart) -> RepoResult<Cart> {
        let u = CartUpdateDiesel::from(update_cart.clone());
        use super::schema::carts::dsl::{id, carts};

        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;

        async_pool::run(move || {
            diesel::update(carts).filter(id.eq(id)).set(u).execute(&conn)
        }).await.map_err(|v| DieselRepoError::from(v).into_inner())?;
        self.find(id).await
    }

    async fn delete(&self, user_id: &u16, delete_cart: &DeleteCart) -> RepoResult<()> {
        let u = DeleteCartDiesel::from(delete_cart.clone());
        use super::schema::carts::dsl::{id, carts};
        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;

        async_pool::run(move || {
            diesel::update(carts).filter(user_id.eq(user_id)).set(u).execute(&conn)
        }).await.map_err(|v| DieselRepoError::from(v).into_inner())?
    }

    async fn create(&self, cart: &Cart) -> RepoResult<()> {
        let u = CartDiesel::from(cart.clone());
        use super::schema::carts::dsl::carts;

        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;
        async_pool::run(move || diesel::insert_into(carts).values(u).execute(&conn))
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())?
    }

    async fn order(&self, destroy: bool) -> RepoResult<Cart> {
        use super::schema::carts::dsl::{destroy, carts};

        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;
        async_pool::run(move || {
            carts.filter(destroy.eq(destroy)).first::<CartDiesel>(&conn)
        })
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())
        .map(|v| -> Cart {v.into()})
    }
}