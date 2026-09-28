-- Accounts live in user-service. users.id now holds the user-service id of
-- each chat participant (no longer AUTO_INCREMENT) and the duplicated
-- username/email columns are dropped; read them from user-service.
--
-- Existing rows use local ids. Before running, fill legacy_user_map with
-- each old id's user-service id (e.g. matched by email):
--   CREATE TABLE legacy_user_map (legacy_id BIGINT PRIMARY KEY, user_id BIGINT NOT NULL);
-- Every reference is remapped in one statement per table, so FKs stay valid.

-- On a fresh database the map is simply empty.
CREATE TABLE IF NOT EXISTS legacy_user_map (legacy_id BIGINT PRIMARY KEY, user_id BIGINT NOT NULL);

SET FOREIGN_KEY_CHECKS = 0;

UPDATE users u JOIN legacy_user_map m ON m.legacy_id = u.id SET u.id = m.user_id;
UPDATE users u JOIN legacy_user_map m ON m.legacy_id = u.manager_id SET u.manager_id = m.user_id;
UPDATE users u JOIN legacy_user_map m ON m.legacy_id = u.advisor_id SET u.advisor_id = m.user_id;
UPDATE conversations c JOIN legacy_user_map m ON m.legacy_id = c.user1_id SET c.user1_id = m.user_id;
UPDATE conversations c JOIN legacy_user_map m ON m.legacy_id = c.user2_id SET c.user2_id = m.user_id;
UPDATE messages x JOIN legacy_user_map m ON m.legacy_id = x.sender_id SET x.sender_id = m.user_id;

SET FOREIGN_KEY_CHECKS = 1;

ALTER TABLE users
    MODIFY id BIGINT NOT NULL,
    DROP INDEX username,
    DROP INDEX email,
    DROP COLUMN username,
    DROP COLUMN email;

DROP TABLE legacy_user_map;
