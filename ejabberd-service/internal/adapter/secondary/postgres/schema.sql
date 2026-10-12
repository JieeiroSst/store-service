CREATE TABLE IF NOT EXISTS rosterusers (
    username text NOT NULL,
    server_host text NOT NULL,
    jid text NOT NULL,
    nick text NOT NULL,
    subscription character(1) NOT NULL,
    ask character(1) NOT NULL,
    askmessage text NOT NULL,
    server character(1) NOT NULL,
    subscribe text NOT NULL,
    "type" text,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS i_rosteru_sh_user_jid ON rosterusers USING btree (server_host, username, jid);
CREATE INDEX IF NOT EXISTS i_rosteru_sh_jid ON rosterusers USING btree (server_host, jid);

CREATE TABLE IF NOT EXISTS rostergroups (
    username text NOT NULL,
    server_host text NOT NULL,
    jid text NOT NULL,
    grp text NOT NULL
);

CREATE INDEX IF NOT EXISTS i_rosterg_sh_user_jid ON rostergroups USING btree (server_host, username, jid);

CREATE TABLE IF NOT EXISTS roster_version (
    username text NOT NULL,
    server_host text NOT NULL,
    version text NOT NULL,
    PRIMARY KEY (server_host, username)
);

CREATE TABLE IF NOT EXISTS spool (
    username text NOT NULL,
    server_host text NOT NULL,
    xml text NOT NULL,
    seq BIGSERIAL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS i_spool_sh_username ON spool USING btree (server_host, username);
