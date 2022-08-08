use std::convert::TryFrom;

#[derive(StructOfArray)]
pub struct Media  {
    pub id: u16,
    pub name: String,
    pub url: String,
    pub description: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}