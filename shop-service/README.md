diesel setup --database-url mysql://root:@localhost:3307/db

diesel migration generate create_posts

diesel migration run --database-url mysql://root:@localhost:3307/db

diesel migration redo --database-url mysql://root:@localhost:3307/db
