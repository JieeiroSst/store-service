use async_trait::async_trait;
use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

use crate::core::{QueryParams, RepoResult, ResultPaging};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Media  {
    pub id: String,
    pub name: String,
    pub url: String,
    pub description: String,
    pub destroy: bool,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

#[async_trait]
pub trait MediaRepo: Send + Sync {
    async fn create(&self, media: Media) -> RepoResult<()>;
    async fn find(&self, id: &u16)-> RepoResult<Media>;
}