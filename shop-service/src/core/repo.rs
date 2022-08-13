use serde::{Deserialize, Serialize};

pub type RepoResult<T> = Result<T, super::error::RepoError>;

pub trait QueryParams: Send + Sync {
    fn limmit(&self) -> i64;
    fn offset(&self) -> i64;
}

const DEFAULT_OFFSET: Option<i64> = Some(0);
const DEFAULT_LIMMIT: Option<i64> = Some(25);

#[derive(Debug, Serialize, Deserialize)]
pub struct QueryParamsImpl {
    pub limmit: Option<i64>,
    pub offset: Option<i64>,
}

impl QueryParamsImpl {
    fn default() -> Self {
        Self::new()
    }
}

impl QueryParams for QueryParamsImpl {
    fn limmit(&self) -> i64 {
        self.limmit.or(DEFAULT_LIMMIT).unwrap_or_default()
    }

    fn offset(&self) -> i64 {
        self.limmit.or(DEFAULT_OFFSET).unwrap_or_default()
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ResultPaging<T> {
    pub total: i64,
    pub items: Vec<T>,
}