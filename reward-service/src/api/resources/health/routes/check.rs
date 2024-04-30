use actix_web::{
    get,
    web::Data,
    HttpResponse,
};

use crate::{
    api::lib::AppState,
    domain::{error::DomainError, health},
};

#[utoipa::path(
    get,
    operation_id = "health",
    path = "/health",
    tag = "health",
    responses(
         (status = 200, description = "health"),
    ),
 )]
#[get("/health")]
async fn handler(state: Data<AppState>) -> Result<HttpResponse, DomainError> {
    let result = health::check::execute(state.health_repository.clone()).await?;
    Ok(HttpResponse::Ok().json(result))
}
