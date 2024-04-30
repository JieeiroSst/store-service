use actix_web::{
    delete,
    web::{self, Data},
    HttpResponse,
};
use uuid::Uuid;

use crate::{
    api::lib::AppState,
    domain::{reward, error::DomainError},
};

#[utoipa::path(
    delete,
    operation_id = "delete_reward",
    path = "/reward/{reward_id}",
    tag = "reward",
    params(
        ("reward_id" = Uuid, Path, description = "reward uuid"),
    ),
    responses(
         (status = 204, description = "reward deleted"),
         (status = 400, description = "Invalid reward id",  body = ErrorResponse),
         (status = 404, description = "reward not found",  body = ErrorResponse),
         (status = 409, description = "reward is in use",  body = ErrorResponse),
    ),
 )]
#[delete("/reward/{reward_id}")]
async fn handler(
    state: Data<AppState>,
    param: web::Path<Uuid>,
) -> Result<HttpResponse, DomainError> {
    reward::resources::delete_by_id::execute(
        state.reward_repository.clone(),
        param.to_owned(),
    )
    .await?;
    Ok(HttpResponse::NoContent().finish())
}
