CREATE TABLE referral_links (
  ref_code      VARCHAR(64)   NOT NULL,
  created_at    BIGINT        NOT NULL,
  owner_user_id VARCHAR(128)  NOT NULL,
  channel       VARCHAR(32)   NOT NULL,
  status        VARCHAR(16)   NOT NULL,
  expires_at    BIGINT        NOT NULL,
  deep_link     VARCHAR(512)  NOT NULL,
  platform      VARCHAR(32)   NOT NULL,
  PRIMARY KEY (ref_code),
  KEY idx_owner_created (owner_user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE referral_events (
  event_id      VARCHAR(36)   NOT NULL,
  ref_code      VARCHAR(64)   NOT NULL,
  event_type    VARCHAR(32)   NOT NULL,
  occurred_at   BIGINT        NOT NULL,
  platform      VARCHAR(32)   NOT NULL,
  new_user_id   VARCHAR(128)  NOT NULL DEFAULT '',
  owner_user_id VARCHAR(128)  NOT NULL,
  ip_address    VARCHAR(64)   NOT NULL DEFAULT '',
  device_id     VARCHAR(128)  NOT NULL DEFAULT '',
  user_agent    VARCHAR(512)  NOT NULL DEFAULT '',
  PRIMARY KEY (event_id),
  KEY idx_ref_code_occurred (ref_code, occurred_at),
  KEY idx_new_user_id (new_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE referral_rewards (
  owner_user_id VARCHAR(128)   NOT NULL,
  ref_code      VARCHAR(64)    NOT NULL,
  new_user_id   VARCHAR(128)   NOT NULL DEFAULT '',
  reward_type   VARCHAR(32)    NOT NULL,
  reward_value  DECIMAL(14,2)  NOT NULL,
  status        VARCHAR(16)    NOT NULL,
  created_at    BIGINT         NOT NULL,
  updated_at    BIGINT         NOT NULL,
  PRIMARY KEY (owner_user_id, ref_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_referral_stats (
  user_id          VARCHAR(128)   NOT NULL,
  total_invited    BIGINT         NOT NULL DEFAULT 0,
  total_installed  BIGINT         NOT NULL DEFAULT 0,
  total_rewarded   BIGINT         NOT NULL DEFAULT 0,
  total_reward_amt DECIMAL(14,2)  NOT NULL DEFAULT 0,
  last_active_at   BIGINT         NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE reward_programs (
  program_id  VARCHAR(36)   NOT NULL,
  name        VARCHAR(128)  NOT NULL,
  status      VARCHAR(16)   NOT NULL,
  tiers       JSON          NOT NULL,
  created_at  BIGINT        NOT NULL,
  updated_at  BIGINT        NOT NULL,
  PRIMARY KEY (program_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE referral_attributions (
  owner_user_id VARCHAR(128)  NOT NULL,
  new_user_id   VARCHAR(128)  NOT NULL,
  ref_code      VARCHAR(64)   NOT NULL,
  platform      VARCHAR(32)   NOT NULL,
  device_id     VARCHAR(128)  NOT NULL DEFAULT '',
  attributed_at BIGINT        NOT NULL,
  PRIMARY KEY (owner_user_id, new_user_id),
  KEY idx_new_user_id (new_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
