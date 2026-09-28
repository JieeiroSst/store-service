-- Users now live in user-service: drop the local credentials table. Owner
-- ids of type 'user' are user-service ids (numeric), while 'corporate' ids
-- stay UUIDs, so the polymorphic wallets.owner_id becomes TEXT like
-- vouchers.owner_id already is.
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_corporate;
DROP TABLE IF EXISTS users;

ALTER TABLE wallets ALTER COLUMN owner_id TYPE TEXT USING owner_id::text;
