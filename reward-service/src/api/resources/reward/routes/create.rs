use actix_web::{
    post,
    web::{self, Data},
    HttpResponse,
};

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
    post,
    operation_id = "create_reward",
    path = "/reward",
    tag = "reward",
    request_body = RequestCreateReward,
    responses(
         (status = 201, description = "reward created",  body = ApiResponseReward),
         (status = 400, description = "Invalid payload",  body = ErrorResponse),
    ),
 )]
#[post("/reward")]
async fn handler(
    state: Data<AppState>,
    body: web::Json<dto::RequestCreateReward>,
) -> Result<HttpResponse, DomainError> {
    body.validate()?;

    let reward =
        reward::resources::create::execute(state.reward_repository.clone(), body.0.into())
            .await?;

    let response = ApiResponse::<ResponseReward>::new(vec![reward.into()], None, None, None);

    Ok(HttpResponse::Created().json(response))
}