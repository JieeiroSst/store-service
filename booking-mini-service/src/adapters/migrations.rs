use log::info;

refinery::embed_migrations!("migrations");

pub async fn run(database_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let (mut client, connection) =
        tokio_postgres::connect(database_url, tokio_postgres::NoTls).await?;

    tokio::spawn(async move {
        if let Err(e) = connection.await {
            eprintln!("migration connection error: {e}");
        }
    });

    let report = migrations::runner().run_async(&mut client).await?;
    info!("Applied {} migration(s)", report.applied_migrations().len());

    Ok(())
}
