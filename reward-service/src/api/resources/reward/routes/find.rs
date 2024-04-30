use actix_web::{
    get,
    web::{Data, Query},
    HttpResponse,
};

use validator::Validate;

use crate::{
    api::{
        config,
        lib::AppState,
        resources::reward::dto::{self, ResponseReward},
        utils::response::ApiResponse,
    },
    domain::{reward, error::DomainError},
};

#[utoipa::path(
    get,
    operation_id = "find_reward",
    path = "/reward",
    tag = "reward",
    params(
        dto::RequestFindReward
    ),
    responses(
         (status = 200, description = "reward",  body = ApiResponseReward),
         (status = 204, description = "no content reward"),
         (status = 400, description = "Invalid query parameters",  body = ErrorResponse),
    ),
 )]
#[get("/reward")]
async fn handler(
    state: Data<AppState>,
    query: Query<dto::RequestFindReward>,
) -> Result<HttpResponse, DomainError> {
    query.validate()?;

    let page = query.page.unwrap_or(1);
    let page_size = query
        .page_size
        .unwrap_or(config::get_config().page_size_default);

    let name = query.name.to_owned();

    let result = reward::resources::find::execute(
        state.reward_repository.clone(),
        name,
        page,
        page_size,
    )
    .await?;

    if let Some((reward, count)) = result {
        let response = ApiResponse::<ResponseReward>::new(
            reward.into_iter().map(|i| i.into()).collect(),
            Some(page),
            Some(count),
            Some(page_size),
        );
        return Ok(HttpResponse::Ok().json(response));
    }

    Ok(HttpResponse::NoContent().finish())
}