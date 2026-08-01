use actix_web::{web, App, HttpServer};
use env_logger::Env;
use booking_mini_service::adapters::api::config_routes;
use booking_mini_service::adapters::db::PostgresRepository;
use booking_mini_service::adapters::migrations;
use log::info;
use std::sync::Arc;
use std::io;

#[actix_web::main]
async fn main() -> io::Result<()> {
    // Initialize logger
    env_logger::init_from_env(Env::default().default_filter_or("info"));
    if dotenv::dotenv().is_err() {
        eprintln!("Note: .env file not found or couldn't be loaded");
    }
    info!("Starting hotel booking service");

    // Load config (DATABASE_URL, PORT) from Consul KV
    let config = booking_mini_service::adapters::consul::load_config()
        .await
        .expect("Failed to load config from Consul");

    let db_url = config.database_url;

    migrations::run(&db_url).await.expect("Failed to run database migrations");

    let repo = Arc::new(PostgresRepository::new(&db_url).await.expect("Failed to connect to database"));

    let port = config.port;
    // Start HTTP server
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(repo.clone()))
            .configure(config_routes)
    })
        .bind(("0.0.0.0", port))?
        .run()
        .await
}