ALTER TABLE referral_rewards
  DROP INDEX idx_referral_rewards_unpublished,
  DROP COLUMN published_at;
