use dotenv::dotenv;

use std::{
    env,
    fmt::{self},
};

#[warn(unused_imports)]
pub struct Config {
    pub database_url: String,
    pub port: String,
}

impl Config {
    pub fn get_database_url() -> Result<Self, dotenv::Error> {
        dotenv()?;

        let config = Config{
            database_url: env::var("DATABASE_URL").unwrap(),
            port: env::var("PORT").unwrap(),
        };
    
        Ok(config)
    }
}
