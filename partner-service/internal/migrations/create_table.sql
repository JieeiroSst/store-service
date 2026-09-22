create table partnerships (
    id int NOT NULL PRIMARY KEY,
    project_id int,
    description varchar(255),
    started_on int,
    expires_on int,
);

create table partnerships_partners (
    id int NOT NULL PRIMARY KEY,
    partner_id int,
    partnership_id int,
    content varchar(255),
    joined_on int,
    left_on int,
);

create table partners (
    id int NOT NULL PRIMARY KEY,
    type varchar(255),
    name varchar(255),
    email varchar(255),
    phone varchar(255),
    address varchar(255),
    status varchar(50) DEFAULT 'active',
    score int DEFAULT 0,
    user_id int,
    created_at int,
    updated_at int,
    deleted_at int DEFAULT 0
);

create table projects (
    id int NOT NULL PRIMARY KEY,
    name varchar(255),
);
