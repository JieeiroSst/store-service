use crate::core::RepoError;

use std::fmt;

#[derive(Debug)]
pub struct DieselRepoError(RepoError);

impl DieselRepoError {
    pub fn into_inner(self) -> RepoError {
        self.0
    }
}

impl From<r2d2::Error> for DieselRepoError {
    fn from(error: r2d2::Error) -> DieselRepoError {
        DieselRepoError(RepoError {
            message: error.to_string()
        })
    }
}

impl From<diesel::result::Error> for DieselRepoError {
    fn from(error: diesel::result::Error) -> DieselRepoError {
        DieselRepoError(RepoError {
            message: error.to_string(),
        })
    }
}

impl<T: fmt::Debug> From<super::async_pool::AsyncPoolError<T>> for DieselRepoError {
    fn from(error: super::async_pool::AsyncPoolError<T>) -> DieselRepoError {
        DieselRepoError(RepoError {
            message: error.to_string(),
        })
    }
}