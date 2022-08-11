use std::sync::Arc;

use async_trait::async_trait;
use crate::core::{CommonError, QueryParams, ResultPaging};

use crate::domain::{Media};

#[async_trait]
pub trait MediaService: Send + Sync {

}

pub struct MediaServiceImpl {
    
}


#[async_trait]
impl MediaService for MediaServiceImpl {

}