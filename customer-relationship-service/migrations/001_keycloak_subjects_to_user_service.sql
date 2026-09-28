-- People used to sign in through Keycloak (OIDC); they now sign in through
-- user-service. uploader_id held the Keycloak "sub"; it must hold the
-- user-service user id so "staff may delete their own uploads" still
-- recognises files uploaded before the switch.
--
-- Fill legacy_user_map first, e.g. by matching emails between a Keycloak
-- realm export and user-service:
--   CREATE TABLE legacy_user_map (legacy_sub VARCHAR(255) PRIMARY KEY, user_id VARCHAR(255) NOT NULL);
-- Uploads whose subject has no mapping keep the old value: only managers
-- and admins (who may delete any file) can remove them.

UPDATE crm_contract_file f
JOIN legacy_user_map m ON m.legacy_sub = f.uploader_id
SET f.uploader_id = m.user_id;

-- The audit trail (crm_contract_file_event.actor_id) is left as recorded:
-- rewriting who did something breaks what an audit log is for. actor_name
-- still identifies the person. If you do want old events searchable by
-- user-service id, run this explicitly:
--
-- UPDATE crm_contract_file_event e
-- JOIN legacy_user_map m ON m.legacy_sub = e.actor_id
-- SET e.actor_id = m.user_id;

DROP TABLE legacy_user_map;
