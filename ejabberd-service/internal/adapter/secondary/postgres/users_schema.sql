CREATE TABLE IF NOT EXISTS users (
    username text NOT NULL,
    server_host text NOT NULL,
    "password" text NOT NULL,
    serverkey text NOT NULL DEFAULT '',
    salt text NOT NULL DEFAULT '',
    iterationcount integer NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (server_host, username)
);
