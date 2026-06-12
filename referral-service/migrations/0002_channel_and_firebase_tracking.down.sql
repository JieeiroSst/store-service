DROP INDEX idx_ref_code_channel ON referral_events;

ALTER TABLE referral_attributions
  DROP COLUMN firebase_instance_id;

ALTER TABLE referral_events
  DROP COLUMN channel,
  DROP COLUMN firebase_instance_id;
