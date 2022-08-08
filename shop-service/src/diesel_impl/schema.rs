table! {
    carts (id) {
        id -> Unsigned<Bigint>,
        total -> Integer,
        user_id -> Integer,
        created_at -> Nullable<Datetime>,
        updated_at -> Nullable<Datetime>,
    }
}

table! {
    cart_items (id) {
        id -> Unsigned<Bigint>,
        cart_id -> Integer,
        total -> Integer,
        created_at -> Nullable<Datetime>,
        updated_at -> Nullable<Datetime>,
    }
}

table! {
    medias (id) {
        id -> Unsigned<Bigint>,
        name -> Text,
        url -> Text,
        description -> Text,
        created_at -> Nullable<Datetime>,
        updated_at -> Nullable<Datetime>,
    }
}

table! {
    products (id) {
        id -> Unsigned<Bigint>,
        product_name -> Text,
        description -> Text,
        price -> Integer,
        media_id -> Integer,
        created_at -> Nullable<Datetime>,
        updated_at -> Nullable<Datetime>,
    }
}

allow_tables_to_appear_in_same_query!(
    carts,
    cart_items,
    medias,
    products,
);
