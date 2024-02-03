use fake::{Dummy, Fake};

#[derive(Clone, Dummy, PartialEq, Eq)]
#[readonly::make]
pub struct Poster {
    pub poster_id: String,
    pub title: String,
    pub description: String,
    pub price: u64,
    pub image_url: String,
}

impl Poster {
    pub fn new(
        poster_id: &str,
        title: &str,
        description: &str,
        price: &str,
        image_url: &str,
    ) -> Self {
        Self {
            poster_id: poster_id.to_string(),
            title: title.to_string(),
            description: description.to_string(),
            price: price.to_string(),
            image_url: image_url.to_string(),
        }
    }

    pub fn edit(&mut self, title: &str, description: &str, price: &str, image_url: &str) {
        self.title = title.to_string();
        self.description = description.to_string();
        self.price = price.to_string();
        self.image_url = image_url.to_string();
    }
}
