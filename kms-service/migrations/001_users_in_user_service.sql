-- Users moved to user-service: user references become user-service ids
-- (BIGINT) and the local users table (with password hashes) is dropped.
--
-- Existing rows reference old local UUIDs, which user-service knows nothing
-- about. Before running this, fill legacy_user_map with each old user's
-- user-service id (e.g. matched by email from an export of both sides):
--
--   CREATE TABLE legacy_user_map (legacy_id UUID PRIMARY KEY, user_id BIGINT NOT NULL);
--   INSERT INTO legacy_user_map VALUES ('…old uuid…', 123456), …;
--
-- Keys whose creator has no mapping keep created_by = 0 (visible to admins
-- only); audit entries without a mapping keep actor_id NULL.
BEGIN;

ALTER TABLE keys ADD COLUMN created_by_user BIGINT;
UPDATE keys SET created_by_user = COALESCE(
    (SELECT m.user_id FROM legacy_user_map m WHERE m.legacy_id = keys.created_by), 0);
ALTER TABLE keys DROP COLUMN created_by;
ALTER TABLE keys RENAME COLUMN created_by_user TO created_by;
ALTER TABLE keys ALTER COLUMN created_by SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_keys_created_by ON keys(created_by);

ALTER TABLE audit_logs ADD COLUMN actor_user BIGINT;
UPDATE audit_logs a SET actor_user = m.user_id
FROM legacy_user_map m WHERE m.legacy_id = a.actor_id;
ALTER TABLE audit_logs DROP COLUMN actor_id;
ALTER TABLE audit_logs RENAME COLUMN actor_user TO actor_id;
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);

DROP TABLE IF EXISTS users;
DROP TABLE legacy_user_map;

COMMIT;
