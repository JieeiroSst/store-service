use std::sync::Arc;

use async_trait::async_trait;
use crate::core::{CommonError, QueryParams, ResultPaging};

use crate::domain::{Media, MediaRepo};

#[async_trait]
pub trait MediaService: Send + Sync {
    async fn Create(&self, media: Media) -> RepoResult<()>;
    async fn find(&self, id: &u16)-> RepoResult<Media>;
}

pub struct MediaServiceImpl {
    pub media_repo: Arc<dyn MediaRepo>
}


#[async_trait]
impl MediaService for MediaServiceImpl {

}