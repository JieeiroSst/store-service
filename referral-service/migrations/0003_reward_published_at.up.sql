-- Outbox marker: NULL means the reward event has not reached the broker yet and
-- the relay will keep retrying it. Rewards that already exist were never sent;
-- they are marked published so enabling the relay does not credit them retroactively.
ALTER TABLE referral_rewards
  ADD COLUMN published_at BIGINT NULL,
  ADD INDEX idx_referral_rewards_unpublished (published_at, created_at);

UPDATE referral_rewards SET published_at = updated_at;
