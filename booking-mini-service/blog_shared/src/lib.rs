use serde::{Serialize, Deserialize};
use std::env;

#[derive(Serialize, Deserialize, Debug)]
pub struct Post {
    port: String,
}

impl Post {
    pub fn new(port: String) -> Self {
        Self { port }
    }
}