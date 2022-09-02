table! {
    cart_items (id) {
        id -> Int4,
        cart_id -> Int4,
        total -> Int4,
        amount -> Int4,
        destroy -> Nullable<Bool>,
        created_at -> Nullable<Time>,
        updated_at -> Nullable<Time>,
    }
}

table! {
    carts (id) {
        id -> Int4,
        total -> Int4,
        user_id -> Int4,
        destroy -> Nullable<Bool>,
        created_at -> Nullable<Time>,
        updated_at -> Nullable<Time>,
    }
}

table! {
    medias (id) {
        id -> Int4,
        name -> Text,
        url -> Text,
        description -> Text,
        destroy -> Nullable<Bool>,
        created_at -> Nullable<Time>,
        updated_at -> Nullable<Time>,
    }
}

table! {
    products (id) {
        id -> Int4,
        product_name -> Text,
        description -> Text,
        price -> Int4,
        media_id -> Int4,
        destroy -> Nullable<Bool>,
        created_at -> Nullable<Time>,
        updated_at -> Nullable<Time>,
    }
}

allow_tables_to_appear_in_same_query!(
    cart_items,
    carts,
    medias,
    products,
);
