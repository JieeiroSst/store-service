use actix_web::{
    get,
    web::{self, Data},
    HttpResponse,
};
use uuid::Uuid;

use crate::{
    api::{
        lib::AppState, resources::reward::dto::ResponseReward, utils::response::ApiResponse,
    },
    domain::{reward, error::DomainError},
};

#[utoipa::path(
    get,
    operation_id = "find_reward_by_id",
    path = "/reward/{reward_id}",
    tag = "reward",
    params(
        ("reward_id" = Uuid, Path, description = "reward uuid"),
    ),
    responses(
         (status = 200, description = "reward finded",  body = ApiResponseReward),
         (status = 204, description = "reward no content"),
    ),
 )]
#[get("/reward/{reward_id}")]
async fn handler(
    state: Data<AppState>,
    param: web::Path<Uuid>,
) -> Result<HttpResponse, DomainError> {
    let result = reward::resources::find_id::execute(
        state.reward_repository.clone(),
        param.to_owned(),
    )
    .await?;

    if let Some(reward) = result {
        let response =
            ApiResponse::<ResponseReward>::new(vec![reward.into()], None, None, None);

        return Ok(HttpResponse::Ok().json(response));
    }

    Ok(HttpResponse::NoContent().finish())
}
