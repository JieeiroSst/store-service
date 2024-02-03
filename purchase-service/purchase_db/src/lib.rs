use postgres::{Client, NoTls};

#[allow(unused_imports)]
pub struct DB {
    pub database_url: String,
}

impl DB {
    pub fn new(database_url: String) -> DB{
        DB { 
            database_url: database_url,
        }
    }

    pub fn connect_to_database(&self) -> Result<Client, postgres::Error> {
        let client = Client::connect(self.database_url.as_str(), NoTls)?;
    
        Ok(client)
    }
}
