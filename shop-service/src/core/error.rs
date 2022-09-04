use serde::{Serialize, Deserialize};

use std::fmt;

#[derive(Serialize, Deserialize, Debug)]
pub struct CommonError {
    pub message: String,
    pub code: u32,
}

impl fmt::Display for CommonError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "Error: {}, Code: {}",self.message, self.code)
    }
}

#[derive(Debug)]
pub struct RepoError {
    pub message: String 
}

impl Into<CommonError> for RepoError {
    fn into(self) -> CommonError {
        CommonError {
            message: self.message,
            code: 1,
        }
    }
}