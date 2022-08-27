use std::sync::Arc;

use async_trait::async_trait;
use chrono::NaiveDateTime;
use diesel::prelude::*;

use super::async_pool;
use super::errors::DieselRepoError;
use super::infra;
use super::schema::*;

use crate::core::{QueryParams, RepoResult, ResultPaging};
use crate::domain::media::{Media, MediaRepo};

#[derive(Queryable, Insertable)]
#[table_name = "medias"]
pub struct MediaDiesel {
    pub id: u16,
    pub name: String,
    pub url: String,
    pub description: String,
    pub destroy: bool,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

impl Into<Media> for MediaDiesel {
    fn into(self) -> Media {
        Media {
            id: self.id,
            name: self.name,
            url: self.url,
            description: self.description,
            destroy: self.description,
            created_at: self.created_at,
            updated_at: self.updated_at,
        }
    }
}

impl From<Media> for MediaDiesel {
    fn from(m: Media) -> Self {
        MediaDiesel {
            id: self.id,
            name: self.name,
            url: self.url,
            description: self.description,
            destroy: self.description,
            created_at: self.created_at,
            updated_at: self.updated_at,
        }
    }
}

pub struct MediaDiesel {
    pool: Arc<infra::DBConn>
}

impl MediaDiesel {
    pub fn new(db: Arc<infra::DBConn>) -> Self {
        MediaDiesel {
            pool: db,
        }
    }
}

#[async_trait]
impl MediaRepo for MediaDiesel {
    async fn Create(&self, media: Media) -> RepoResult<()> {
        let u = MediaDiesel::from(media.clone());
        use super::schema::carts::dsl::carts;

        let conn = self.pool.get().map_err(|v| DieselRepoError::from(v).into_inner())?;
        async_pool::run(move || {
            diesel::insert_into(medias).value(u).execute(&conn).await
            .map_err(|v| DieselRepoError::from(v).into_inner())?;
        });
    }

    async fn find(&self, id: &u16)-> RepoResult<Media> {
        use super::schema::medias::dsl::{id, medias};
        let conn = self.pool.get().map_err(|v| DieselRepoError::from::into_inner())?;

        async_pool::run(move || medias.filter((id.eq(id)).first::<MediaDiesel>(&conn)))
        .await
        .map_err(|v| DieselRepoError::from(v).into_inner())
        .map(|v| -> Media {v.into()})
    }
}