use actix_web::{web, App, HttpServer};
use env_logger::Env;
use booking_mini_service::adapters::api::config_routes;
use booking_mini_service::adapters::db::PostgresRepository;
use log::info;
use std::sync::Arc;
use std::env;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize logger
    env_logger::init_from_env(Env::default().default_filter_or("info"));
    if dotenv::dotenv().is_err() {
        eprintln!("Note: .env file not found or couldn't be loaded");
    }
    info!("Starting hotel booking service");

    // Database connection
    let db_url = env::var("DATABASE_URL")
        .unwrap_or_else(|_| "postgres://postgres:postgres@localhost:31000/hotel_booking".to_string());

    let repo = Arc::new(PostgresRepository::new(&db_url).await.expect("Failed to connect to database"));

    let port = env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse::<u16>()
        .expect("PORT must be a valid number");
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