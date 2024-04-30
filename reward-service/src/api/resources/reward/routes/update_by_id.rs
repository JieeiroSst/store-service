use actix_web::{
    put,
    web::{self, Data},
    HttpResponse,
};
use uuid::Uuid;
use validator::Validate;

use crate::{
    api::{
        lib::AppState,
        resources::reward::dto::{self, ResponseReward},
        utils::response::ApiResponse,
    },
    domain::{reward, error::DomainError},
};

#[utoipa::path(
    put,
    operation_id = "update_reward",
    path = "/reward/{reward_id}",
    tag = "reward",
    params(
        ("reward_id" = Uuid, Path, description = "reward uuid"),
    ),
    request_body = RequestUpdateReward,
    responses(
         (status = 200, description = "reward updated",  body = ApiResponseReward),
         (status = 400, description = "Invalid payload",  body = ErrorResponse),
         (status = 404, description = "reward not found",  body = ErrorResponse),
    ),
 )]
#[put("/reward/{reward_id}")]
async fn handler(
    state: Data<AppState>,
    param: web::Path<Uuid>,
    body: web::Json<dto::RequestUpdateReward>,
) -> Result<HttpResponse, DomainError> {
    body.validate()?;

    let reward = reward::resources::update::execute(
        state.reward_repository.clone(),
        param.to_owned(),
        body.0.into(),
    )
    .await?;

    let response = ApiResponse::<ResponseReward>::new(vec![reward.into()], None, None, None);

    Ok(HttpResponse::Ok().json(response))
}