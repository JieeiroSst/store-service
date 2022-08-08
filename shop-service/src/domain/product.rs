use std::convert::TryFrom;

#[derive(StructOfArray)]
pub struct Product {
    pub id: u16,
    pub name: String,
    pub description: String,
    pub price: u16,
    pub media_id: u16,
    pub media: Media,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub trait Product {
    pub fn Create(product: Product) -> Result<Self, Self::Error>;
    pub fn Update(id: u16, product: Product) -> Result<Self, Self::Error>;

}

pub impl Product for Product {
    type Error = ();
    fn Create(product: Product) -> Result<Self, Self::Error> {
        let isCheck: bool;
        if !isCheck {
            Error(())
        }
        Ok(Self(product))
    }
} 