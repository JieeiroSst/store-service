use serde::Deserialize;
use std::env;
use std::error::Error;

#[derive(Debug, Deserialize)]
pub struct AppConfig {
    #[serde(rename = "DATABASE_URL")]
    pub database_url: String,
    #[serde(rename = "PORT")]
    pub port: u16,
}

pub async fn load_config() -> Result<AppConfig, Box<dyn Error>> {
    let host = env::var("HostConsul")?;
    let key = env::var("KeyConsul")?;

    let url = format!("{}/v1/kv/{}?raw", host.trim_end_matches('/'), key);

    let response = reqwest::get(&url).await?.error_for_status()?;
    let config = response.json::<AppConfig>().await?;

    Ok(config)
}
