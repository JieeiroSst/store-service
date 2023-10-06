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
);

create table projects (
    id int NOT NULL PRIMARY KEY,
    name varchar(255),
);
