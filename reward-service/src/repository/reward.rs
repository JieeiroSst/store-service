use std::sync::Arc;

use async_trait::async_trait;
use deadpool_postgres::Pool;

use tokio_postgres::{types::ToSql, Row};
use uuid::Uuid;

use crate::domain::{
    reward::{
        model::{RewardCreateModel, RewardModel, RewardUpdateModel},
        repository::RewardRepository,
    },
    error::DomainError,
};

const QUERY_FIND_REWARD: &str = "
    select 
        id, 
        name,
        description,
        points,
        created_at,
        updated_at
    from rewards 
";

const QUERY_FIND_REWARD_BY_ID: &str = "
    select 
        id, 
        name,
        description,
        points,
        created_at,
        updated_at
    from rewards 
    where 
        id = $1;
";

const QUERY_INSERT_REWARD: &str = "
    insert into rewards 
    (name, description, points) 
    values ($1,$2,$3)
    returning
        id, 
        name,
        description,
        points,
        created_at,
        updated_at
";

const QUERY_UPDATE_REWARD: &str = "
    update rewards
    set name = $2,
        description = $3,
        points = $3
    where 
        id = $1
    returning
        id, 
        name,
        description,
        points,
        created_at,
        updated_at
";

const QUERY_DELETE_REWARD_BY_ID: &str = "
    delete from
        rewards 
    where
        id = $1;
";

pub struct PgRewardRepository {
    pool: Arc<Pool>,
}
impl PgRewardRepository {
    pub fn new(pool: Arc<Pool>) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl RewardRepository for PgRewardRepository {
    async fn find(
        &self,
        name: &Option<String>,
        page: &u32,
        page_size: &u32,
    ) -> Result<Option<(Vec<RewardModel>, u32)>, DomainError> {
        let client = self.pool.get().await?;

        let mut queries: Vec<String> = vec![];
        let mut params: Vec<&(dyn ToSql + Sync)> = Vec::new();

        if let Some(name) = name {
            queries.push(format!(
                "reward.name like '%' || ${} || '%'",
                params.len() + 1
            ));
            params.push(name);
        }
        let mut query = String::from(QUERY_FIND_REWARD);
        if !queries.is_empty() {
            query = format!("{} where {}", query, queries.join(" and "));
        }

        let offset = page_size * (page - 1);
        query = format!("{query} limit {page_size} offset {offset}");

        let stmt = client.prepare(&query).await?;
        let result = client.query(&stmt, &params[..]).await?;

        if !result.is_empty() {
            let count: u32 = result.first().unwrap().get("count");

            let categories: Vec<RewardModel> = result.iter().map(|row| row.into()).collect();

            return Ok(Some((categories, count)));
        }
    }

    async fn find_by_id(&self, id: &Uuid) -> Result<Option<RewardModel>, DomainError> {
        let client = self.pool.get().await?;
        let stmt = client.prepare(QUERY_FIND_REWARD_BY_ID).await?;

        if let Some(result) = client.query_opt(&stmt, &[id]).await? {
            return Ok(Some((&result).into()));
        }

        return Ok(None);
    }

    async fn insert(
        &self,
        reward_create_model: &RewardCreateModel,
    ) -> Result<RewardModel, DomainError> {
        let client = self.pool.get().await?;
        let stmt = client.prepare(QUERY_INSERT_REWARD).await?;
        let result = &client
            .query_one(
                &stmt,
                &[
                    &reward_create_model.id,
                    &reward_create_model.name,
                    &reward_create_model.description,
                    &reward_create_model.points,
                ],
            )
            .await?;

        Ok(result.into())
    }

    async fn update_by_id(
        &self,
        id: &Uuid,
        reward_create_model: &RewardUpdateModel,
    ) -> Result<RewardModel, DomainError> {
        let client = self.pool.get().await?;
        let stmt = client.prepare(QUERY_UPDATE_REWARD).await?;
        let result = &client
            .query_one(
                &stmt,
                &[
                    id,
                    &reward_create_model.name,
                    &reward_create_model.description,
                    &reward_create_model.points,
                ],
            )
            .await?;

        Ok(result.into())
    }    

    async fn delete_by_id(&self, id: &Uuid) -> Result<(), DomainError> {
        let client = self.pool.get().await?;
        let stmt = client.prepare(QUERY_DELETE_REWARD_BY_ID).await?;
        client.execute(&stmt, &[id]).await?;
        Ok(())
    }
}

impl From<&Row> for RewardModel {
    fn from(row: &Row) -> Self {
        Self {
            id: row.get("id"),
            name: row.get("name"),
            description: row.get("description"),
            points: row.get("points"),
            created_at: row.get("created_at"),
            updated_at: row.get("updated_at"),
        }
    }
}