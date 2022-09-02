extern crate dotenv;

use std::env;
use diesel;
use diesel::r2d2::ConnectionManager;

pub type Pool<T> = r2d2::Pool<ConnectionManager<T>>;
pub type PGPool = Pool<diesel::pg::PgConnection>;

#[cfg(feature = "postgres")]
pub type DBConn = PGPool;

pub fn db_pool() -> DBConn {
    let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");
    println!("Using Database {}", database_url);
    let manager = ConnectionManager::<diesel::pg::PgConnection>::new(database_url);
    Pool::builder()
        .build(manager)
        .expect("Failed to create pool")
}