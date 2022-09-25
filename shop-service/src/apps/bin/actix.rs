#[actix_web::main]
async fn main() -> std::io::Result<()> {
    shopservice::apps::actix::server::serve().await
}