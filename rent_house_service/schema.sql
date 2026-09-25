create table if not exists homestay
(
    id           bigint generated always as identity primary key,
    name         text not null,
    description  text,
    type         integer,
    status       integer,
    phone_number text,

    address      text,
    ward_id      integer,
    district_id  integer,
    province_id  integer,

    images text[],

    guests       smallint,
    bedrooms     smallint,
    bathrooms    smallint,

    extra_data   jsonb,
    version      bigint,
    created_at   timestamp with time zone,
    created_by   bigint,
    updated_at   timestamp with time zone,
    updated_by   bigint
);

create table if not exists booking
(
    id            bigint generated always as identity primary key,
    user_id       bigint   not null, -- owned by user-service, no FK
    homestay_id   bigint   not null,
    checkin_date  date     not null,
    checkout_date date     not null,
    guests        smallint not null,
    status        smallint not null,

    currency      text     not null,
    subtotal      numeric(18, 4),
    discount      numeric(18, 4),
    total_amount  numeric  not null,
    price_detail  jsonb,

    note          text,
    request_id    text     not null unique,

    version       bigint,
    extra_data    jsonb,
    created_at    timestamp with time zone,
    created_by    bigint,
    updated_at    timestamp with time zone,
    updated_by    bigint
);

create table if not exists homestay_availability
(
    homestay_id bigint not null,
    date        date   not null,
    price       numeric,
    status      smallint,
    primary key (homestay_id, date)
);


create table if not exists amenity
(
    id   integer generated always as identity primary key,
    name text not null,
    icon text not null
);

create table if not exists homestay_amenity
(
    homestay_id bigint  not null constraint homestay_amenity_homestay_id_fk references homestay,
    amenity_id  integer not null constraint homestay_amenity_amenity_id_fk references amenity,
    primary key (homestay_id, amenity_id)
);

create table if not exists ward
(
    id          integer generated always as identity primary key,
    ward_name   text not null,
    district_id integer
);


create table if not exists district
(
    id            integer generated always as identity primary key,
    district_name text not null,
    province_id   integer
);

create table if not exists province
(
    id            integer generated always as identity primary key,
    province_name text not null,
    country_id    integer
);

create index if not exists booking_user_id_idx on booking (user_id, created_at desc);
create index if not exists booking_homestay_id_idx on booking (homestay_id, checkin_date);
create index if not exists homestay_location_idx on homestay (province_id, district_id, ward_id);

-- Payment state. Nights of a booking stay held (homestay_availability.status = booked)
-- while it waits for payment (booking.status = 3) and are released when the hold expires.
alter table booking add column if not exists payment_method text;
alter table booking add column if not exists payment_ref text;
alter table booking add column if not exists paid_at timestamp with time zone;
alter table booking add column if not exists expires_at timestamp with time zone;
create index if not exists booking_pending_idx on booking (expires_at) where status = 3;

-- Rental models. Nightly stays (model = day) book nights from the calendar above; week, month
-- and year are long-term leases priced from the manager's rate plan, one rate per model.
create table if not exists homestay_rate
(
    homestay_id bigint   not null references homestay,
    model       text     not null check (model in ('day', 'week', 'month', 'year')),
    price       numeric(14, 2) not null check (price >= 0),
    currency    text     not null,
    min_periods integer  not null default 1 check (min_periods >= 1),
    active      boolean  not null default true,
    updated_by  bigint,
    updated_at  timestamp with time zone,
    primary key (homestay_id, model)
);

-- A lease holds its nights from start_date (inclusive) to end_date (exclusive).
-- status: 1 pending (waiting for the first invoice), 2 active, 3 ended, 4 cancelled.
create table if not exists lease
(
    id           bigint generated always as identity primary key,
    user_id      bigint   not null, -- owned by user-service, no FK
    homestay_id  bigint   not null references homestay,
    model        text     not null check (model in ('week', 'month', 'year')),
    start_date   date     not null,
    end_date     date     not null,
    periods      integer  not null check (periods >= 1),
    -- day of month rent is collected (month model), ISO weekday 1-7 (week model), 0 for year
    billing_day  smallint not null default 0,
    currency     text     not null,
    rent         numeric(14, 2) not null,
    status       smallint not null,
    note         text,
    request_id   text     not null unique,
    expires_at   timestamp with time zone,
    version      bigint,
    created_at   timestamp with time zone,
    updated_at   timestamp with time zone
);
create index if not exists lease_user_idx on lease (user_id, id desc);
create index if not exists lease_pending_idx on lease (expires_at) where status = 1;

-- One invoice per rent period. status: 1 unpaid, 2 paid, 3 void, 4 refunded (paid, then the
-- lease ended before the period began).
create table if not exists rent_invoice
(
    id             bigint generated always as identity primary key,
    lease_id       bigint   not null references lease,
    period_no      integer  not null,
    period_start   date     not null,
    period_end     date     not null,
    due_date       date     not null,
    amount         numeric(14, 2) not null,
    status         smallint not null default 1,
    payment_method text,
    payment_ref    text,
    paid_at        timestamp with time zone,
    version        bigint   not null default 1,
    unique (lease_id, period_no)
);
create index if not exists rent_invoice_due_idx on rent_invoice (due_date) where status = 1;

-- Reviews come from guests who really stayed: a paid booking that has ended, or a lease.
create table if not exists review
(
    id          bigint generated always as identity primary key,
    homestay_id bigint   not null references homestay,
    user_id     bigint   not null,
    booking_id  bigint references booking,
    lease_id    bigint references lease,
    rating      smallint not null check (rating between 1 and 5),
    comment     text,
    created_at  timestamp with time zone not null default now(),
    check ((booking_id is null) <> (lease_id is null))
);
create unique index if not exists review_booking_uq on review (booking_id) where booking_id is not null;
create unique index if not exists review_lease_uq on review (lease_id) where lease_id is not null;
create index if not exists review_homestay_idx on review (homestay_id, id desc);

create table if not exists wishlist
(
    user_id     bigint not null,
    homestay_id bigint not null references homestay,
    created_at  timestamp with time zone not null default now(),
    primary key (user_id, homestay_id)
);

-- The original numeric(12, 6) held at most 999999.999999, which overflows on a two-night VND stay.
alter table booking alter column subtotal type numeric(18, 4);
alter table booking alter column discount type numeric(18, 4);

-- Each homestay is paid into its own wallet in payment-wallet-service (set by the manager).
alter table homestay add column if not exists wallet_id text;

-- Anyone can register a homestay, hotel or rental house; it goes live once the platform admin
-- approves it. status: 1 active, 2 inactive (owner paused it), 3 pending approval, 4 rejected.
-- type: 1 homestay, 2 hotel, 3 rental house.
alter table homestay add column if not exists review_note text;
create index if not exists homestay_owner_idx on homestay (created_by, id desc);
create index if not exists homestay_status_idx on homestay (status, id);
