ALTER TABLE referral_links
  ADD COLUMN utm_source   VARCHAR(128) NOT NULL DEFAULT '' AFTER platform,
  ADD COLUMN utm_medium   VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_source,
  ADD COLUMN utm_campaign VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_medium,
  ADD COLUMN utm_term     VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_campaign,
  ADD COLUMN utm_content  VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_term;

ALTER TABLE referral_events
  ADD COLUMN utm_source   VARCHAR(128) NOT NULL DEFAULT '' AFTER firebase_instance_id,
  ADD COLUMN utm_medium   VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_source,
  ADD COLUMN utm_campaign VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_medium,
  ADD COLUMN utm_term     VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_campaign,
  ADD COLUMN utm_content  VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_term;

ALTER TABLE referral_attributions
  ADD COLUMN utm_source   VARCHAR(128) NOT NULL DEFAULT '' AFTER firebase_instance_id,
  ADD COLUMN utm_medium   VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_source,
  ADD COLUMN utm_campaign VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_medium,
  ADD COLUMN utm_term     VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_campaign,
  ADD COLUMN utm_content  VARCHAR(128) NOT NULL DEFAULT '' AFTER utm_term;

CREATE INDEX idx_utm_campaign_occurred ON referral_events (utm_campaign, occurred_at);
CREATE INDEX idx_owner_occurred ON referral_events (owner_user_id, occurred_at);
