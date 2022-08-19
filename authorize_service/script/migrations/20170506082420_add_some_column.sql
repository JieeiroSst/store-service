-- +goose Up
CREATE TABLE test (
    id int NOT NULL,
    title text,
    body text,
    PRIMARY KEY(id)
);

-- +goose Down
DROP TABLE test;
