use actix_web::{web, App, HttpServer};
use env_logger::Env;
use sqlx::postgres::PgPoolOptions;
use std::env;

mod adapters;
mod application;
mod domain;
mod ports;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize logger
    env_logger::init_from_env(Env::default().default_filter_or("info"));
    log::info!("Starting hotel management system");

    // Load database connection from environment or use default
    let database_url = env::var("DATABASE_URL")
        .unwrap_or_else(|_| "postgres://postgres:postgres@localhost/hotel_db".to_string());

    // Create connection pool
    let pool = PgPoolOptions::new()
        .max_connections(5)
        .connect(&database_url)
        .await
        .expect("Failed to create pool");

    // Run database migrations
    sqlx::migrate!("./migrations")
        .run(&pool)
        .await
        .expect("Failed to run migrations");

    // Setup HTTP server with routes
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(pool.clone()))
            .service(adapters::api::room::configure())
            .service(adapters::api::booking::configure())
            .service(adapters::api::passenger::configure())
    })
        .bind(("127.0.0.1", 8080))?
        .run()
        .await
}