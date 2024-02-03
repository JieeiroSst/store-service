use redis::{Client, Commands};

#[warn(unreachable_patterns)]
pub struct Cache {
    pub dns: String,
}

impl Cache {
    pub fn new(dns: String) -> Cache {
        Cache { dns }
    }
    pub fn call_redis(
        &self,
        command: &str,
        key: &str,
        value: Option<&str>,
    ) -> Result<String, redis::RedisError> {
        let client = Client::open(self.dns.as_str())?;
        let mut con = client.get_connection()?;

        match command {
            _get => {
                let result: String = con.get(key)?;
                Ok(result)
            }
            _set => {
                con.set(key, value)?;
                Ok("OK".to_string()) // Redis returns "OK" on successful SET
            }
        }
    }
}

// let result = call_redis("GET", "my_key", None).unwrap();
// println!("Value: {}", result);

// call_redis("SET", "new_key", Some("my_value")).unwrap();
