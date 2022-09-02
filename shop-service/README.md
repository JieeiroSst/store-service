diesel setup --database-url "postgres://root:postgres@localhost/web_shop"

diesel migration run  --database-url "postgres://root:postgres@localhost/web_shop"

diesel migration repo  --database-url "postgres://root:postgres@localhost/web_shop"