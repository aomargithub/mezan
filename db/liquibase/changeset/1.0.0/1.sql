--liquibase formatted sql

--changeset aomar:1
create table users
(
  id serial not null PRIMARY KEY,
  name varchar(500) not null,
  email varchar(500) not null UNIQUE,
  last_updated_at  timestamp null,
  created_at timestamp not null,
  hashed_password CHAR(60) NULL
);
--rollback drop table users;


--changeset aomar:2
create table mezanis
(
  id serial not null PRIMARY KEY,
  name varchar(500) not null,
  created_at timestamp not null,
  last_updated_at  timestamp null,
  creator_id integer not null references users(id),
  settled_percent float not null,
  total_amount float not null
);
--rollback drop table mezanis;

--changeset aomar:3
create table expenses
(
  id serial not null PRIMARY KEY,
  name varchar(500) not null,
  last_updated_at  timestamp null,
  created_at timestamp not null,
  settled_percent float not null,
  total_amount float not null,
  creator_id integer not null references users(id),
  mezani_id integer not null references mezanis(id),
  receipt varchar(500) null
); 
--rollback drop table expenses;

--changeset aomar:4
create table expense_items
(
  id serial not null PRIMARY KEY,
  name varchar(500) not null,
  last_updated_at  timestamp null,
  created_at timestamp not null,
  settled_percent float not null,
  amount float not null,
  expense_id integer not null references expenses(id),
  mezani_id integer not null references mezanis(id),
  creator_id integer not null references users(id) 
); 
--rollback drop table expense_items;

--changeset aomar:5
create table payments
(
  id serial not null PRIMARY KEY,
  created_at timestamp not null,
  settled_percent float not null,
  amount float not null,
  expense_item_id integer not null references expense_items(id),
  mezani_id integer not null references mezanis(id),
  creator_id integer not null references users(id) 
); 
--rollback drop table payments;


--changeset aomar:6
CREATE TABLE sessions (
    token CHAR(43) PRIMARY KEY,
    data bytea NOT NULL,
    expiry timestamp NOT NULL
);
--rollback drop table sessions;

CREATE INDEX sessions_expiry_idx ON sessions (expiry);