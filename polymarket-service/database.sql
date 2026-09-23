-- polymarket-service schema. Money never lives in a wallet row here: users
-- deposit into a treasury wallet on payment-wallet-service and the exchange keeps
-- their trading cash in balances. Prices and cash are integers in minor
-- currency units, a winning share pays markets.share_value.
--
-- Applied on startup by internal/infrastructure/database.applySchema, which splits
-- the file on every semicolon character, so keep semicolons out of comments.
-- Indexes are declared inline because MySQL has no CREATE INDEX IF NOT EXISTS.

CREATE TABLE IF NOT EXISTS events (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    slug        VARCHAR(160) NOT NULL,
    title       VARCHAR(512) NOT NULL,
    description TEXT,
    category    VARCHAR(64) NOT NULL DEFAULT '',
    tags        TEXT,
    image_url   VARCHAR(512) NOT NULL DEFAULT '',
    featured    TINYINT(1) NOT NULL DEFAULT 0,
    neg_risk    TINYINT(1) NOT NULL DEFAULT 0,
    status      VARCHAR(16) NOT NULL DEFAULT 'open',
    end_date    DATETIME NOT NULL,
    volume      BIGINT NOT NULL DEFAULT 0,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_events_slug (slug),
    INDEX idx_events_status_volume (status, volume),
    INDEX idx_events_category (category, status),
    INDEX idx_events_end_date (end_date)
);

CREATE TABLE IF NOT EXISTS markets (
    id               BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id         BIGINT NOT NULL,
    slug             VARCHAR(190) NOT NULL,
    question         VARCHAR(512) NOT NULL,
    group_item_title VARCHAR(255) NOT NULL DEFAULT '',
    description      TEXT,
    share_value      BIGINT NOT NULL DEFAULT 100,
    min_order_size   BIGINT NOT NULL DEFAULT 1,
    end_time         DATETIME NOT NULL,
    status           VARCHAR(16) NOT NULL DEFAULT 'open',
    proposed_outcome VARCHAR(8) NOT NULL DEFAULT '',
    dispute_deadline DATETIME NULL,
    disputed_by      VARCHAR(64) NOT NULL DEFAULT '',
    dispute_reason   TEXT,
    resolved_outcome VARCHAR(8) NOT NULL DEFAULT '',
    reward_pool      BIGINT NOT NULL DEFAULT 0,
    reward_max_spread BIGINT NOT NULL DEFAULT 0,
    reward_min_size  BIGINT NOT NULL DEFAULT 0,
    dispute_bond     BIGINT NOT NULL DEFAULT 0,
    bond_settled     TINYINT(1) NOT NULL DEFAULT 0,
    best_bid         BIGINT NOT NULL DEFAULT 0,
    best_ask         BIGINT NOT NULL DEFAULT 0,
    last_price       BIGINT NOT NULL DEFAULT 0,
    volume           BIGINT NOT NULL DEFAULT 0,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_markets_slug (slug),
    INDEX idx_markets_event (event_id),
    INDEX idx_markets_status_volume (status, volume)
);

-- yes_price and book_side place every order on one YES-denominated book: buy YES
-- and sell NO are bids, sell YES and buy NO are asks. The index serves the
-- matcher's price-time ordering.
CREATE TABLE IF NOT EXISTS orders (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    market_id       BIGINT NOT NULL,
    user_id         VARCHAR(64) NOT NULL,
    client_order_id VARCHAR(64) NOT NULL DEFAULT '',
    outcome         VARCHAR(8) NOT NULL,
    side            VARCHAR(8) NOT NULL,
    type            VARCHAR(8) NOT NULL,
    time_in_force   VARCHAR(8) NOT NULL,
    price           BIGINT NOT NULL,
    size            BIGINT NOT NULL,
    filled          BIGINT NOT NULL DEFAULT 0,
    budget          BIGINT NOT NULL DEFAULT 0,
    locked_cash     BIGINT NOT NULL DEFAULT 0,
    fee_reserve     BIGINT NOT NULL DEFAULT 0,
    fee             BIGINT NOT NULL DEFAULT 0,
    filled_cash     BIGINT NOT NULL DEFAULT 0,
    yes_price       BIGINT NOT NULL,
    book_side       VARCHAR(4) NOT NULL,
    status          VARCHAR(12) NOT NULL DEFAULT 'open',
    expires_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_orders_book (market_id, status, book_side, yes_price, id),
    INDEX idx_orders_user (user_id, status, id),
    INDEX idx_orders_client (user_id, client_order_id)
);

CREATE TABLE IF NOT EXISTS trades (
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    market_id      BIGINT NOT NULL,
    yes_price      BIGINT NOT NULL,
    size           BIGINT NOT NULL,
    kind           VARCHAR(8) NOT NULL,
    maker_order_id BIGINT NOT NULL,
    taker_order_id BIGINT NOT NULL,
    maker_user_id  VARCHAR(64) NOT NULL,
    taker_user_id  VARCHAR(64) NOT NULL,
    maker_outcome  VARCHAR(8) NOT NULL,
    maker_side     VARCHAR(8) NOT NULL,
    taker_outcome  VARCHAR(8) NOT NULL,
    taker_side     VARCHAR(8) NOT NULL,
    taker_fee      BIGINT NOT NULL DEFAULT 0,
    maker_rebate   BIGINT NOT NULL DEFAULT 0,
    referral_fee   BIGINT NOT NULL DEFAULT 0,
    referrer_user_id VARCHAR(64) NOT NULL DEFAULT '',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_trades_market (market_id, id),
    INDEX idx_trades_maker (maker_user_id, id),
    INDEX idx_trades_taker (taker_user_id, id),
    INDEX idx_trades_created (created_at),
    INDEX idx_trades_referrer (referrer_user_id, id)
);

CREATE TABLE IF NOT EXISTS balances (
    user_id    VARCHAR(64) NOT NULL PRIMARY KEY,
    available  BIGINT NOT NULL DEFAULT 0,
    locked     BIGINT NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id     VARCHAR(64) NOT NULL,
    type        VARCHAR(24) NOT NULL,
    amount      BIGINT NOT NULL,
    market_id   BIGINT NOT NULL DEFAULT 0,
    transfer_id VARCHAR(64) NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ledger_user (user_id, id)
);

CREATE TABLE IF NOT EXISTS positions (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    market_id     BIGINT NOT NULL,
    user_id       VARCHAR(64) NOT NULL,
    outcome       VARCHAR(8) NOT NULL,
    shares        BIGINT NOT NULL DEFAULT 0,
    locked_shares BIGINT NOT NULL DEFAULT 0,
    cost_basis    BIGINT NOT NULL DEFAULT 0,
    realized_pnl  BIGINT NOT NULL DEFAULT 0,
    settled       TINYINT(1) NOT NULL DEFAULT 0,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_positions_market_user_outcome (market_id, user_id, outcome),
    INDEX idx_positions_user (user_id, settled),
    INDEX idx_positions_holders (market_id, outcome, shares)
);

CREATE TABLE IF NOT EXISTS profiles (
    user_id    VARCHAR(64) NOT NULL PRIMARY KEY,
    username   VARCHAR(30) NOT NULL,
    bio        VARCHAR(500) NOT NULL DEFAULT '',
    avatar_url VARCHAR(512) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_profiles_username (username)
);

CREATE TABLE IF NOT EXISTS comments (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id   BIGINT NOT NULL,
    parent_id  BIGINT NOT NULL DEFAULT 0,
    user_id    VARCHAR(64) NOT NULL,
    body       TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_comments_event (event_id, id),
    INDEX idx_comments_parent (parent_id)
);

CREATE TABLE IF NOT EXISTS comment_likes (
    comment_id BIGINT NOT NULL,
    user_id    VARCHAR(64) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (comment_id, user_id)
);

CREATE TABLE IF NOT EXISTS bookmarks (
    user_id    VARCHAR(64) NOT NULL,
    event_id   BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, event_id)
);

-- Realised profit and loss, one row per event, so profit can be ranked over any
-- time window. Open positions are not included.
CREATE TABLE IF NOT EXISTS pnl_entries (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id    VARCHAR(64) NOT NULL,
    market_id  BIGINT NOT NULL DEFAULT 0,
    kind       VARCHAR(16) NOT NULL,
    amount     BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pnl_user (user_id, created_at),
    INDEX idx_pnl_created (created_at, user_id)
);

-- The exchange's own insert-only books: fee revenue and funding, liquidity
-- rewards paid out (negative), and dispute bonds held. Revenue and bond are
-- separate buckets so a refund can never eat into reward money.
CREATE TABLE IF NOT EXISTS exchange_entries (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    bucket     VARCHAR(16) NOT NULL,
    kind       VARCHAR(24) NOT NULL,
    amount     BIGINT NOT NULL,
    market_id  BIGINT NOT NULL DEFAULT 0,
    user_id    VARCHAR(64) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_exchange_bucket (bucket)
);

-- One row per sampling epoch. Whichever replica inserts it first does that
-- epoch's payout, so several replicas can run the reward job safely.
CREATE TABLE IF NOT EXISTS reward_epochs (
    epoch_start DATETIME NOT NULL PRIMARY KEY,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reward_payouts (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    epoch_start DATETIME NOT NULL,
    market_id   BIGINT NOT NULL,
    user_id     VARCHAR(64) NOT NULL,
    score       DOUBLE NOT NULL,
    amount      BIGINT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_reward_payout (epoch_start, market_id, user_id),
    INDEX idx_reward_user (user_id, id)
);

CREATE TABLE IF NOT EXISTS referral_codes (
    user_id    VARCHAR(64) NOT NULL PRIMARY KEY,
    ref_code   VARCHAR(64) NOT NULL,
    deep_link  VARCHAR(512) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS referrals (
    referee_user_id  VARCHAR(64) NOT NULL PRIMARY KEY,
    referrer_user_id VARCHAR(64) NOT NULL,
    ref_code         VARCHAR(64) NOT NULL,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_referrals_referrer (referrer_user_id)
);
