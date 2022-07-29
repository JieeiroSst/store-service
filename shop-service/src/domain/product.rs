#[derive(StructOfArray)]
pub struct Product {
    pub id: u16,
    pub name: String,
    pub description: String,
    pub price: u16,
    pub media_id: u16,
    pub media: Media
}