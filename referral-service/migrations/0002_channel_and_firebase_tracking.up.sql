ALTER TABLE referral_events
  ADD COLUMN channel VARCHAR(32) NOT NULL DEFAULT '' AFTER platform,
  ADD COLUMN firebase_instance_id VARCHAR(128) NOT NULL DEFAULT '' AFTER user_agent;

ALTER TABLE referral_attributions
  ADD COLUMN firebase_instance_id VARCHAR(128) NOT NULL DEFAULT '' AFTER attributed_at;

CREATE INDEX idx_ref_code_channel ON referral_events (ref_code, channel);
