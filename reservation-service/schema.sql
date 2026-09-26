-- reservation-service owns hotels, their room types and rooms, the per-day inventory and rate of
-- each room type, and reservations. Guests are not stored here: guest_id is a user id in user-service.

-- Hotels are 5-star only. status: 1 active, 2 inactive (paused by its manager), 3 pending
-- verification by the platform admin, 4 rejected.
create table if not exists hotel
(
    id                bigint generated always as identity primary key,
    name              text     not null,
    description       text,
    stars             smallint not null check (stars = 5),
    city              text     not null, -- "location": where guests search for it
    address           text     not null,
    phone_number      text     not null,
    email             text,
    images            text[],
    amenities         text[],
    check_in_time     text     not null default '14:00',
    check_out_time    text     not null default '12:00',
    currency          text     not null,
    -- a paid reservation can be cancelled with a full refund until this long before check-in
    free_cancel_hours integer  not null default 48 check (free_cancel_hours >= 0),
    status            smallint not null,
    review_note       text,
    owner_id          bigint   not null, -- the manager (user in user-service) who registered it
    wallet_id         text,              -- wallet in payment-wallet-service that receives payments
    version           bigint,
    created_at        timestamp with time zone,
    updated_at        timestamp with time zone
);
create index if not exists hotel_owner_idx on hotel (owner_id, id desc);
create index if not exists hotel_status_idx on hotel (status, id);
create index if not exists hotel_city_idx on hotel (lower(city));

create table if not exists room_type
(
    id          bigint generated always as identity primary key,
    hotel_id    bigint   not null references hotel,
    name        text     not null,
    description text,
    capacity    smallint not null check (capacity >= 1), -- guests per room
    bed         text,
    size_m2     integer,
    view        text,
    amenities   text[],
    active      boolean  not null default true,
    unique (hotel_id, name)
);

create table if not exists room
(
    id           bigint generated always as identity primary key,
    hotel_id     bigint  not null references hotel,
    room_type_id bigint  not null references room_type,
    name         text    not null,
    floor        integer,
    is_available boolean not null default true,
    unique (hotel_id, name)
);
create index if not exists room_type_idx on room (room_type_id);

-- One row per room type per night: how many rooms exist, how many are reserved, and the rate.
-- A night without a rate is not for sale.
create table if not exists room_type_inventory
(
    hotel_id        bigint  not null references hotel,
    room_type_id    bigint  not null references room_type,
    date            date    not null,
    total_inventory integer not null check (total_inventory >= 0),
    total_reserved  integer not null default 0,
    rate            numeric(14, 2) check (rate >= 0),
    primary key (room_type_id, date),
    check (total_reserved >= 0 and total_reserved <= total_inventory)
);
create index if not exists inventory_hotel_date_idx on room_type_inventory (hotel_id, date);

-- Extra services a hotel sells with a stay (airport transfer, spa, late check-out, ...).
-- unit: 'stay' = charged once, 'night' = charged per night.
create table if not exists hotel_service
(
    id          bigint generated always as identity primary key,
    hotel_id    bigint not null references hotel,
    name        text   not null,
    description text,
    price       numeric(14, 2) not null check (price >= 0),
    unit        text   not null check (unit in ('stay', 'night')),
    active      boolean not null default true,
    unique (hotel_id, name)
);

-- status: 1 pending, 2 paid, 3 canceled, 4 rejected, 5 refunded.
--   pending -> paid | canceled | rejected;  paid -> refunded
create table if not exists reservation
(
    id               bigint generated always as identity primary key,
    hotel_id         bigint   not null references hotel,
    room_type_id     bigint   not null references room_type,
    guest_id         bigint   not null, -- user in user-service, no FK
    start_date       date     not null,
    end_date         date     not null, -- check-out day, exclusive
    rooms            integer  not null check (rooms >= 1),
    adults           integer  not null check (adults >= 1),
    children         integer  not null default 0 check (children >= 0),
    status           smallint not null,
    currency         text     not null,
    room_total       numeric(18, 4) not null,
    addons_total     numeric(18, 4) not null default 0,
    total_amount     numeric(18, 4) not null,
    special_requests text,
    request_id       text     not null unique,
    payment_method   text,
    payment_ref      text,
    paid_at          timestamp with time zone,
    expires_at       timestamp with time zone,
    status_note      text,
    version          bigint,
    created_at       timestamp with time zone,
    updated_at       timestamp with time zone,
    check (end_date > start_date)
);
create index if not exists reservation_guest_idx on reservation (guest_id, id desc);
create index if not exists reservation_hotel_idx on reservation (hotel_id, id desc);
create index if not exists reservation_pending_idx on reservation (expires_at) where status = 1;

create table if not exists reservation_service
(
    reservation_id bigint  not null references reservation,
    service_id     bigint  not null references hotel_service,
    name           text    not null,
    quantity       integer not null check (quantity >= 1),
    unit_price     numeric(14, 2) not null,
    amount         numeric(18, 4) not null,
    primary key (reservation_id, service_id)
);

-- Reviews come from guests who really stayed: a paid reservation whose check-out has passed.
create table if not exists review
(
    id             bigint generated always as identity primary key,
    hotel_id       bigint   not null references hotel,
    user_id        bigint   not null,
    reservation_id bigint   not null unique references reservation,
    rating         smallint not null check (rating between 1 and 5),
    comment        text,
    created_at     timestamp with time zone not null default now()
);
create index if not exists review_hotel_idx on review (hotel_id, id desc);

-- ---- cancellation policy, fees and refunds
-- The tiers of a hotel's cancellation policy, e.g. [{"hours_before":48,"fee_percent":0},{"hours_before":24,"fee_percent":50}]:
-- cancelling at least 48h before check-in is free, at least 24h before costs 50%, later is non-refundable.
-- NULL means "free until free_cancel_hours before check-in, then non-refundable".
alter table hotel add column if not exists cancellation_tiers jsonb;

alter table reservation add column if not exists fee_amount numeric(18, 4) not null default 0;
alter table reservation add column if not exists refund_amount numeric(18, 4) not null default 0;

-- ---- promo codes
create table if not exists promotion
(
    id          bigint generated always as identity primary key,
    hotel_id    bigint   not null references hotel,
    code        text     not null,
    description text,
    percent_off smallint check (percent_off between 1 and 100),
    amount_off  numeric(14, 2) check (amount_off > 0),
    min_nights  integer  not null default 1 check (min_nights >= 1),
    max_uses    integer check (max_uses >= 1), -- NULL = unlimited
    used_count  integer  not null default 0 check (used_count >= 0),
    valid_from  date,
    valid_to    date,
    active      boolean  not null default true,
    check ((percent_off is null) <> (amount_off is null))
);
create unique index if not exists promotion_code_uq on promotion (hotel_id, lower(code));

alter table reservation add column if not exists promo_id bigint references promotion;
alter table reservation add column if not exists promo_code text;
alter table reservation add column if not exists discount_amount numeric(18, 4) not null default 0;

-- ---- the stay itself. stay_status: 0 not arrived, 1 checked in, 2 checked out, 3 no-show
alter table reservation add column if not exists stay_status smallint not null default 0;
alter table reservation add column if not exists checked_in_at timestamp with time zone;
alter table reservation add column if not exists checked_out_at timestamp with time zone;
alter table reservation add column if not exists reminded_at timestamp with time zone;

-- ---- what happened to a reservation, and who did it (actor 0 = the system)
create table if not exists reservation_event
(
    id             bigint generated always as identity primary key,
    reservation_id bigint   not null references reservation,
    at             timestamp with time zone not null default now(),
    actor          bigint   not null default 0,
    event          text     not null,
    from_status    smallint,
    to_status      smallint,
    note           text
);
create index if not exists reservation_event_idx on reservation_event (reservation_id, id);

-- ---- waiting list: guests who wanted rooms that were sold out. status: 1 waiting, 2 offered (a reservation was
-- created for them, see reservation_id), 3 cancelled by the guest
create table if not exists waitlist
(
    id             bigint generated always as identity primary key,
    hotel_id       bigint   not null references hotel,
    room_type_id   bigint   not null references room_type,
    guest_id       bigint   not null,
    start_date     date     not null,
    end_date       date     not null,
    rooms          integer  not null check (rooms >= 1),
    adults         integer  not null check (adults >= 1),
    children       integer  not null default 0,
    status         smallint not null default 1,
    reservation_id bigint references reservation,
    created_at     timestamp with time zone not null default now(),
    offered_at     timestamp with time zone,
    check (end_date > start_date)
);
create index if not exists waitlist_waiting_idx on waitlist (room_type_id, id) where status = 1;
create index if not exists waitlist_guest_idx on waitlist (guest_id, id desc);

-- ---- notifications: the inbox is the source of truth; pushed_at says whether notification-service got it
create table if not exists notification
(
    id            bigint generated always as identity primary key,
    user_id       bigint   not null,
    kind          text     not null,
    title         text     not null,
    body          text     not null,
    reservation_id bigint,
    hotel_id      bigint,
    created_at    timestamp with time zone not null default now(),
    read_at       timestamp with time zone,
    pushed_at     timestamp with time zone,
    push_attempts integer  not null default 0
);
create index if not exists notification_user_idx on notification (user_id, id desc);
create index if not exists notification_unpushed_idx on notification (id) where pushed_at is null and push_attempts < 5;

-- ---- guests' favourite hotels
create table if not exists wishlist
(
    user_id    bigint not null,
    hotel_id   bigint not null references hotel,
    created_at timestamp with time zone not null default now(),
    primary key (user_id, hotel_id)
);

-- One notification per event and person: a retry of the code that raises it cannot notify twice.
create unique index if not exists notification_event_uq on notification (user_id, kind, reservation_id) where reservation_id is not null;
create index if not exists reservation_remind_idx on reservation (start_date) where status = 2 and reminded_at is null;

-- A replica claims a batch of notifications before pushing them, so several replicas never push the same one.
alter table notification add column if not exists claimed_until timestamp with time zone;
