DROP INDEX idx_owner_occurred ON referral_events;
DROP INDEX idx_utm_campaign_occurred ON referral_events;

ALTER TABLE referral_attributions
  DROP COLUMN utm_source,
  DROP COLUMN utm_medium,
  DROP COLUMN utm_campaign,
  DROP COLUMN utm_term,
  DROP COLUMN utm_content;

ALTER TABLE referral_events
  DROP COLUMN utm_source,
  DROP COLUMN utm_medium,
  DROP COLUMN utm_campaign,
  DROP COLUMN utm_term,
  DROP COLUMN utm_content;

ALTER TABLE referral_links
  DROP COLUMN utm_source,
  DROP COLUMN utm_medium,
  DROP COLUMN utm_campaign,
  DROP COLUMN utm_term,
  DROP COLUMN utm_content;
