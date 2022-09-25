// @generated automatically by Diesel CLI.

diesel::table! {
    cart_items (id) {
        id -> Text,
        cart_id -> Text,
        total -> Int4,
        amount -> Int4,
        destroy -> Nullable<Bool>,
        created_at -> Timestamp,
        updated_at -> Timestamp,
    }
}

diesel::table! {
    carts (id) {
        id -> Text,
        total -> Int4,
        user_id -> Text,
        destroy -> Nullable<Bool>,
        created_at -> Timestamp,
        updated_at -> Timestamp,
    }
}

diesel::table! {
    medias (id) {
        id -> Text,
        name -> Text,
        url -> Text,
        description -> Text,
        destroy -> Nullable<Bool>,
        created_at -> Timestamp,
        updated_at -> Timestamp,
    }
}

diesel::table! {
    products (id) {
        id -> Text,
        product_name -> Text,
        description -> Text,
        price -> Int4,
        media_id -> Text,
        destroy -> Nullable<Bool>,
        created_at -> Timestamp,
        updated_at -> Timestamp,
    }
}

diesel::allow_tables_to_appear_in_same_query!(
    cart_items,
    carts,
    medias,
    products,
);
