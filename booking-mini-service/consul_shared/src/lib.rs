use consulrs::client::{ConsulClient, ConsulClientSettingsBuilder};
use serde::{Serialize, Deserialize};
use consulrs::kv;

#[derive(Serialize, Deserialize, Debug)]
pub struct Config {
    port: String
}

impl Config {
    pub async fn setup_config() -> Self{
        let client = ConsulClient::new(
            ConsulClientSettingsBuilder::default()
                .address("https://127.0.0.1:8200")
                .build()
                .unwrap()
        ).unwrap();

        let mut res_port = kv::read(&client, "port", None).await.unwrap();
        let port: String = res_port.response.pop().unwrap().value.unwrap().try_into().unwrap();
        Self { port }
    }
}