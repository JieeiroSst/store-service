use async_trait::async_trait;
use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

use crate::core::{QueryParams, RepoResult, ResultPaging};

#[derive(StructOfArray,Debug, Clone, Serialize, Deserialize)]
pub struct CartItem {
    pub id:  u16,
    pub cart_id: u16,
    pub product: Product,
    pub total: u16,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[async_trait]
pub trait CartItemRepo: Send + Sync {
    
}