-- ticket-service sells tickets to events (concerts, shows, sports, workshops). Buyers and organizers are not stored
-- here: user_id / organizer_id are user ids in user-service. Money is an integer in minor units (VND has none).

-- status: 1 draft, 2 pending review by the platform admin, 3 published, 4 rejected, 5 cancelled.
create table if not exists event
(
    id                  bigint generated always as identity primary key,
    organizer_id        bigint      not null,
    title               text        not null,
    description         text,
    category            text        not null,
    city                text        not null,
    venue               text        not null,
    address             text,
    banner_url          text,
    starts_at           timestamptz not null,
    ends_at             timestamptz not null check (ends_at > starts_at),
    currency            text        not null default 'VND',
    status              smallint    not null,
    review_note         text,
    featured            boolean     not null default false,
    transferable        boolean     not null default true, -- may holders pass their tickets on?
    wallet_id           text,                -- wallet in payment-wallet-service that receives the sales
    -- a paid order can be refunded until this long before the event starts; 0 means tickets are non-refundable
    refund_cutoff_hours integer     not null default 0 check (refund_cutoff_hours >= 0),
    version             bigint      not null default 1,
    created_at          timestamptz not null default now(),
    updated_at          timestamptz not null default now()
);
create index if not exists event_organizer_idx on event (organizer_id, id desc);
create index if not exists event_public_idx on event (starts_at) where status = 3;
create index if not exists event_status_idx on event (status, id);
create index if not exists event_city_idx on event (lower(city));
create index if not exists event_category_idx on event (category) where status = 3;

-- A ticket type is a tier or zone of an event (VIP, GA, Early bird, ...).
--   total      tickets the organizer put on sale
--   available  tickets nobody holds or owns; taken by one atomic UPDATE per purchase attempt
--   sold       tickets of paid orders
-- so total - available - sold are held by unpaid orders. The CHECKs are the last line of defence: whatever the
-- application does, the database cannot sell a ticket twice.
create table if not exists ticket_type
(
    id             bigint generated always as identity primary key,
    event_id       bigint   not null references event,
    name           text     not null,
    description    text,
    price          bigint   not null check (price >= 0),
    total          integer  not null check (total >= 0),
    available      integer  not null check (available >= 0),
    sold           integer  not null default 0 check (sold >= 0),
    min_per_order  smallint not null default 1 check (min_per_order >= 1),
    max_per_order  smallint not null default 10 check (max_per_order >= min_per_order),
    max_per_user   smallint not null default 0 check (max_per_user >= 0), -- across all of one user's live orders; 0 = no cap
    sale_starts_at timestamptz,
    sale_ends_at   timestamptz,
    seated         boolean  not null default false,
    active         boolean  not null default true,
    sort_order     integer  not null default 0,
    check (available + sold <= total)
);
create index if not exists ticket_type_event_idx on ticket_type (event_id, sort_order, id);

-- Seats of a seated ticket type. status: 1 available, 2 held by a pending order, 3 sold.
create table if not exists seat
(
    id             bigint generated always as identity primary key,
    ticket_type_id bigint   not null references ticket_type,
    section        text     not null default '',
    row_label      text     not null,
    number         integer  not null,
    status         smallint not null default 1,
    order_id       bigint
);
create index if not exists seat_type_idx on seat (ticket_type_id, status);
create index if not exists seat_order_idx on seat (order_id) where order_id is not null;

-- kind: 1 percent off, 2 fixed amount off.
create table if not exists promotion
(
    id          bigint generated always as identity primary key,
    event_id    bigint      not null references event,
    code        text        not null,
    kind        smallint    not null check (kind in (1, 2)),
    value       bigint      not null check (value > 0),
    max_uses    integer     not null default 0 check (max_uses >= 0), -- 0 = unlimited
    used        integer     not null default 0 check (used >= 0),
    min_tickets integer     not null default 1 check (min_tickets >= 1),
    valid_from  timestamptz,
    valid_to    timestamptz,
    active      boolean     not null default true,
    created_at  timestamptz not null default now(),
    unique (event_id, code),
    check (max_uses = 0 or used <= max_uses)
);

-- status: 1 pending (tickets held until expires_at), 2 paid, 3 expired, 4 cancelled, 5 refunded.
create table if not exists ticket_order
(
    id             bigint generated always as identity primary key,
    user_id        bigint      not null,
    event_id       bigint      not null references event,
    status         smallint    not null,
    currency       text        not null,
    subtotal       bigint      not null check (subtotal >= 0),
    discount       bigint      not null default 0 check (discount >= 0),
    total          bigint      not null check (total >= 0),
    promo_id       bigint,
    promo_code     text,
    buyer_name     text        not null,
    buyer_email    text        not null,
    buyer_phone    text,
    request_id     text        not null, -- client-chosen: a retry of the same request returns the same order
    expires_at     timestamptz not null,
    payment_method text,
    payment_ref    text,
    paid_at        timestamptz,
    refund_amount  bigint      not null default 0,
    status_note    text,
    reminded_at    timestamptz,
    version        bigint      not null default 1,
    created_at     timestamptz not null default now(),
    updated_at     timestamptz not null default now(),
    unique (user_id, request_id)
);
create index if not exists order_user_idx on ticket_order (user_id, id desc);
create index if not exists order_user_status_idx on ticket_order (user_id, status);
create index if not exists order_event_idx on ticket_order (event_id, id desc);
create index if not exists order_expiry_idx on ticket_order (expires_at) where status = 1;
create index if not exists order_reminder_idx on ticket_order (event_id) where status = 2 and reminded_at is null;
create index if not exists order_event_open_idx on ticket_order (event_id) where status in (1, 2);

create table if not exists order_item
(
    id             bigint generated always as identity primary key,
    order_id       bigint  not null references ticket_order,
    ticket_type_id bigint  not null references ticket_type,
    name           text    not null,
    quantity       integer not null check (quantity >= 1),
    unit_price     bigint  not null,
    seat_ids       bigint[]
);
create index if not exists order_item_order_idx on order_item (order_id);
create index if not exists order_item_type_idx on order_item (ticket_type_id);

-- One row per admission. code is what the QR shows. Its life (see ticket_event for the history):
--   status: 1 valid, 2 void (refunded, cancelled or revoked), 3 used (checked in), 4 expired (the event ended unused),
--           5 transferring (offered to someone else and not scannable until they answer)
-- holder_id is who holds the ticket now: the buyer at first, someone else after a transfer.
create table if not exists ticket
(
    id             bigint generated always as identity primary key,
    order_id       bigint      not null references ticket_order,
    event_id       bigint      not null references event,
    ticket_type_id bigint      not null references ticket_type,
    seat_id        bigint,
    code           text        not null unique,
    status         smallint    not null default 1,
    checked_in_at  timestamptz,
    checked_in_by  bigint,
    holder_id      bigint,
    holder_name    text,
    holder_email   text,
    transfer_count smallint    not null default 0,
    created_at     timestamptz not null default now()
);
create index if not exists ticket_order_idx on ticket (order_id);
create index if not exists ticket_event_idx on ticket (event_id, id);

create table if not exists wishlist
(
    user_id    bigint      not null,
    event_id   bigint      not null references event,
    created_at timestamptz not null default now(),
    primary key (user_id, event_id)
);
create index if not exists wishlist_user_idx on wishlist (user_id, created_at desc);

-- The inbox is the source of truth; pushed_at says whether notification-service got it.
create table if not exists notification
(
    id            bigint generated always as identity primary key,
    user_id       bigint      not null,
    kind          text        not null,
    title         text        not null,
    body          text        not null,
    order_id      bigint,
    event_id      bigint,
    created_at    timestamptz not null default now(),
    read_at       timestamptz,
    pushed_at     timestamptz,
    push_attempts integer     not null default 0,
    claimed_until timestamptz
);
create index if not exists notification_user_idx on notification (user_id, id desc);
create index if not exists notification_unpushed_idx on notification (id) where pushed_at is null and push_attempts < 5;
-- One notification per event and person: a retry of the code that raises it cannot notify twice.
create unique index if not exists notification_once_uq on notification (user_id, kind, order_id) where order_id is not null;

-- ---- ticket lifecycle

-- Every change to a ticket, oldest first: who did what, and when.
create table if not exists ticket_event
(
    id          bigint generated always as identity primary key,
    ticket_id   bigint      not null references ticket,
    event       text        not null, -- issued, checked_in, check_in_reverted, transfer_offered, transfer_accepted,
                                      -- transfer_declined, transfer_cancelled, transfer_expired, reissued, holder_changed, voided, expired
    from_status smallint,
    to_status   smallint,
    actor       bigint,               -- user id; null for the system
    note        text,
    created_at  timestamptz not null default now()
);
create index if not exists ticket_event_ticket_idx on ticket_event (ticket_id, id);

-- status: 1 pending, 2 accepted, 3 declined, 4 cancelled by the sender, 5 expired.
create table if not exists ticket_transfer
(
    id          bigint generated always as identity primary key,
    ticket_id   bigint      not null references ticket,
    from_user   bigint      not null,
    to_user     bigint,               -- the recipient is named by user id, by e-mail, or both
    to_email    text,
    status      smallint    not null default 1,
    message     text,
    created_at  timestamptz not null default now(),
    expires_at  timestamptz not null,
    resolved_at timestamptz,
    check (to_user is not null or to_email is not null)
);
-- a ticket has at most one open offer
create unique index if not exists ticket_transfer_open_uq on ticket_transfer (ticket_id) where status = 1;
create index if not exists ticket_transfer_from_idx on ticket_transfer (from_user, id desc);
create index if not exists ticket_transfer_to_idx on ticket_transfer (to_user, id desc) where status = 1;
create index if not exists ticket_transfer_email_idx on ticket_transfer (lower(to_email)) where status = 1;
create index if not exists ticket_transfer_expiry_idx on ticket_transfer (expires_at) where status = 1;

-- Users who may scan tickets at the gate of an event, besides its organizer.
create table if not exists event_staff
(
    event_id   bigint      not null references event,
    user_id    bigint      not null,
    added_by   bigint      not null,
    created_at timestamptz not null default now(),
    primary key (event_id, user_id)
);
create index if not exists event_staff_user_idx on event_staff (user_id);

-- "Tell me when tickets are back": people waiting on a sold-out ticket type, oldest first.
create table if not exists waitlist
(
    ticket_type_id bigint      not null references ticket_type,
    user_id        bigint      not null,
    created_at     timestamptz not null default now(),
    primary key (ticket_type_id, user_id)
);
create index if not exists waitlist_user_idx on waitlist (user_id, created_at desc);
create index if not exists waitlist_queue_idx on waitlist (ticket_type_id, created_at);

-- ---- upgrades of databases created before these columns existed

alter table event add column if not exists transferable boolean not null default true;
alter table ticket add column if not exists holder_id bigint;
alter table ticket add column if not exists holder_name text;
alter table ticket add column if not exists holder_email text;
alter table ticket add column if not exists transfer_count smallint not null default 0;
alter table ticket_order add column if not exists invited_by bigint; -- set on complimentary orders: the organizer who invited
-- Both fixes below match nothing once done; their partial indexes are then empty, so re-running them at every start costs nothing.
create index if not exists ticket_no_holder_idx on ticket (id) where holder_id is null;
create index if not exists ticket_legacy_used_idx on ticket (id) where status = 1 and checked_in_at is not null;
update ticket t set holder_id = o.user_id, holder_name = o.buyer_name, holder_email = o.buyer_email
    from ticket_order o where o.id = t.order_id and t.holder_id is null;
update ticket set status = 3 where status = 1 and checked_in_at is not null;
create index if not exists ticket_holder_idx on ticket (holder_id, id desc);
create index if not exists ticket_live_idx on ticket (event_id) where status in (1, 5);

-- ---- graphical seat maps: where each seat is drawn, and the outline of the venue

alter table seat add column if not exists x double precision;
alter table seat add column if not exists y double precision;
alter table ticket_type add column if not exists layout jsonb; -- stage and section shapes: see domain.SeatMapLayout

-- ---- showtimes: events created from one another share a series

alter table event add column if not exists series_id bigint; -- the id of the first event of the series
create index if not exists event_series_idx on event (series_id, starts_at) where series_id is not null;

-- ---- e-mail: a notification may also be an e-mail, sent by notification-service from a queue kept here

alter table notification add column if not exists email_to text;
alter table notification add column if not exists email_template text;
alter table notification add column if not exists email_data jsonb;
alter table notification add column if not exists emailed_at timestamptz;
alter table notification add column if not exists email_attempts integer not null default 0;
alter table notification add column if not exists email_claimed_until timestamptz;
create index if not exists notification_unemailed_idx on notification (id) where email_template is not null and emailed_at is null and email_attempts < 5;

-- ---- resale: a holder sells a ticket to another buyer at up to its face value

alter table event add column if not exists resale_cap_percent integer not null default 0; -- 0: no resale; 100: up to face value

-- status: 1 open, 2 processing (a buyer has claimed it and is paying), 3 sold, 4 cancelled.
-- While a listing is open or processing its ticket is "listed" (status 6): it cannot be scanned or transferred.
create table if not exists ticket_resale
(
    id          bigint generated always as identity primary key,
    ticket_id   bigint      not null references ticket,
    seller_id   bigint      not null,
    buyer_id    bigint,
    price       bigint      not null check (price > 0),
    fee         bigint      not null default 0,     -- the platform's share of price, set when sold
    currency    text        not null,
    status      smallint    not null default 1,
    payment_ref text,                               -- the wallet transfers of the buyer's payment, once made
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now()
);
create unique index if not exists ticket_resale_live_uq on ticket_resale (ticket_id) where status in (1, 2);
create index if not exists ticket_resale_open_idx on ticket_resale (id desc) where status = 1;
create index if not exists ticket_resale_seller_idx on ticket_resale (seller_id, id desc);
create index if not exists ticket_resale_stale_idx on ticket_resale (updated_at) where status = 2;

-- ---- venues: the seating of a hall, described once and used for many events

create table if not exists venue
(
    id          bigint generated always as identity primary key,
    owner_id    bigint      not null,
    name        text        not null,
    city        text,
    address     text,
    description text,
    shared      boolean     not null default false, -- in the library every organizer may copy from (set by an admin)
    layout      jsonb       not null,               -- the blueprint: see domain.VenueMap
    seats       integer     not null,
    sections    integer     not null,
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now()
);
create index if not exists venue_owner_idx on venue (owner_id, id desc);
create index if not exists venue_shared_idx on venue (id desc) where shared;

-- the venue an event's seat map came from, and the drawing as it was when it was applied (later edits of the venue
-- do not move seats of an event that is on sale)
alter table event add column if not exists venue_id bigint;
alter table event add column if not exists seat_map jsonb;

alter table seat add column if not exists accessible boolean not null default false;
-- a seat is found by its section, row and number, so sections of one ticket type may each start at row A
alter table seat drop constraint if exists seat_ticket_type_id_row_label_number_key;
create unique index if not exists seat_position_uq on seat (ticket_type_id, section, row_label, number);

-- ---- documents: the PDFs of a paid order (invoice, e-tickets), kept in upload-service under the buyer's id

create table if not exists order_document
(
    order_id   bigint      not null references ticket_order,
    kind       text        not null,            -- invoice | tickets
    file_id    text        not null,            -- the file in upload-service
    file_name  text        not null,
    size       bigint      not null default 0,
    created_at timestamptz not null default now(),
    primary key (order_id, kind)
);
alter table ticket_order add column if not exists documents_at timestamptz; -- set once the order's documents exist
create index if not exists order_documents_pending_idx on ticket_order (id) where status = 2 and documents_at is null;

-- ---- showtimes: one event, several sessions (Friday 8pm, Saturday 8pm, ...)

-- status: 1 scheduled, 2 cancelled. Every ticket type belongs to one session, and so does every order and ticket: what
-- happens by date (a refund cutoff, a transfer, a scan, an expiry, a reminder) is decided by the session's dates, not the
-- event's. event.starts_at / ends_at are kept as the first start and the last end of the scheduled sessions.
create table if not exists event_session
(
    id            bigint generated always as identity primary key,
    event_id      bigint      not null references event,
    starts_at     timestamptz not null,
    ends_at       timestamptz not null check (ends_at > starts_at),
    label         text,
    status        smallint    not null default 1,
    cancel_reason text,
    created_at    timestamptz not null default now()
);
create index if not exists event_session_event_idx on event_session (event_id, starts_at);
create index if not exists event_session_open_idx on event_session (ends_at) where status = 1;

alter table ticket_type add column if not exists session_id bigint;
alter table ticket_order add column if not exists session_id bigint;
alter table ticket add column if not exists session_id bigint;
-- Databases from before sessions: every event gets the one session it always had, and everything points at it. Each statement
-- matches nothing once done, and its partial index is then empty, so running them at every start costs nothing.
insert into event_session (event_id, starts_at, ends_at)
    select e.id, e.starts_at, e.ends_at from event e where not exists (select 1 from event_session s where s.event_id = e.id);
create index if not exists ticket_type_no_session_idx on ticket_type (id) where session_id is null;
create index if not exists ticket_order_no_session_idx on ticket_order (id) where session_id is null;
create index if not exists ticket_no_session_idx on ticket (id) where session_id is null;
update ticket_type t set session_id = (select s.id from event_session s where s.event_id = t.event_id order by s.starts_at, s.id limit 1) where t.session_id is null;
update ticket_order o set session_id = (select tt.session_id from order_item i join ticket_type tt on tt.id = i.ticket_type_id where i.order_id = o.id limit 1) where o.session_id is null;
update ticket k set session_id = (select tt.session_id from ticket_type tt where tt.id = k.ticket_type_id) where k.session_id is null;
create index if not exists ticket_type_session_idx on ticket_type (session_id);
create index if not exists ticket_session_idx on ticket (session_id) where status in (1, 5, 6);
create index if not exists order_session_idx on ticket_order (session_id) where status in (1, 2);

-- a ticket type's name is unique within its session (every showtime of an event has its own VIP)
alter table ticket_type drop constraint if exists ticket_type_event_id_name_key;
create unique index if not exists ticket_type_session_name_uq on ticket_type (session_id, name);
