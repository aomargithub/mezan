--liquibase formatted sql

--changeset aomar:1
create table users
(
    id              serial       not null PRIMARY KEY,
    name            varchar(500) not null,
    email           varchar(500) not null UNIQUE,
    last_updated_at timestamp    null,
    created_at      timestamp    not null,
    hashed_password CHAR(60)     NULL
);
--rollback drop table users;


--changeset aomar:2
create table mezanis
(
    id               serial       not null PRIMARY KEY,
    name             varchar(500) not null,
    created_at       timestamp    not null,
    last_updated_at  timestamp    null,
    creator_id       integer      not null references users (id),
    allocated_amount float        not null DEFAULT 0.00,
    total_amount     float        not null DEFAULT 0.00,
    share_id         varchar(500) not null unique,
    has_expenses     boolean      not null DEFAULT true,
    CONSTRAINT unique_mezani_name_per_creator UNIQUE (creator_id, name)
);
--rollback drop table mezanis;

--changeset aomar:3
create table mezani_membership
(
    id         serial    not null PRIMARY KEY,
    created_at timestamp not null,
    member_id  integer   not null references users (id),
    mezani_id  integer   not null references mezanis (id),
    CONSTRAINT unique_member_per_mezani UNIQUE (member_id, mezani_id)
);
--rollback drop table mezani_membership;

--changeset aomar:4
create table expenses
(
    id               serial       not null PRIMARY KEY,
    name             varchar(500) not null,
    last_updated_at  timestamp    null,
    created_at       timestamp    not null,
    allocated_amount float        not null DEFAULT 0.00,
    total_amount     float        not null,
    creator_id       integer      not null references users (id),
    mezani_id        integer      not null references mezanis (id),
    receipt          varchar(500) null,
    has_items        boolean      not null DEFAULT true,
    CONSTRAINT unique_expense_name_per_mezani UNIQUE (name, mezani_id)
);
--rollback drop table expenses;

--changeset aomar:5
CREATE INDEX expenses_mezani_idx ON expenses (mezani_id);
--rollback drop index expenses_mezani_idx;

--changeset aomar:6
create table expense_items
(
    id               serial       not null PRIMARY KEY,
    name             varchar(500) not null,
    last_updated_at  timestamp    null,
    created_at       timestamp    not null,
    allocated_amount float        not null DEFAULT 0.00,
    amount           float        not null,
    total_amount     float        not null,
    expense_id       integer      not null references expenses (id),
    mezani_id        integer      not null references mezanis (id),
    creator_id       integer      not null references users (id),
    quantity         float        not null,
    CONSTRAINT unique_item_name_per_expense UNIQUE (name, expense_id)
);
--rollback drop table expense_items;

--changeset aomar:7
CREATE INDEX expense_items_expense_idx ON expenses (id);
--rollback drop index expense_items_expense_idx;

--changeset aomar:8
CREATE TYPE share_type AS ENUM ('PERCENTAGE', 'EXACT');
--rollback DROP TYPE share_type;

--changeset aomar:9
create table mezani_shares
(
    id             serial     not null PRIMARY KEY,
    created_at     timestamp  not null,
    last_updated_at  timestamp    null,
    share          float      not null,
    amount         float      not null,
    share_type     share_type not null,
    mezani_id      integer    not null references mezanis (id),
    participant_id integer    not null references users (id),
    CONSTRAINT unique_participant_per_mezani_share UNIQUE (mezani_id, participant_id)
);
--rollback drop table mezani_shares;

--changeset aomar:10
CREATE INDEX mezani_shares_mezani_idx ON mezanis (id);
--rollback drop index mezani_shares_mezani_idx;

--changeset aomar:11
create table expense_shares
(
    id             serial     not null PRIMARY KEY,
    created_at     timestamp  not null,
    last_updated_at  timestamp    null,
    share          float      not null,
    amount         float      not null,
    share_type     share_type not null,
    expense_id     integer    not null references expenses (id),
    mezani_id      integer    not null references mezanis (id),
    participant_id integer    not null references users (id),
    CONSTRAINT unique_participant_per_expense_share UNIQUE (expense_id, participant_id)
);
--rollback drop table expense_shares;

--changeset aomar:12
CREATE INDEX expense_shares_expense_idx ON expenses (id);
--rollback drop index expense_shares_expense_idx;

--changeset aomar:13
create table expense_item_shares
(
    id              serial     not null PRIMARY KEY,
    created_at      timestamp  not null,
    last_updated_at  timestamp    null,
    share           float      not null,
    amount          float      not null,
    share_type      share_type not null,
    expense_item_id integer    not null references expense_items (id),
    expense_id      integer    not null references expenses (id),
    mezani_id       integer    not null references mezanis (id),
    participant_id  integer    not null references users (id),
    CONSTRAINT unique_participant_per_expense_item_share UNIQUE (expense_item_id, participant_id)
);
--rollback drop table expense_item_shares;

--changeset aomar:14
CREATE INDEX expense_item_shares_expense_item_idx ON expense_items (id);
--rollback drop index expense_item_shares_expense_item_idx;

--changeset aomar:15
CREATE TABLE sessions
(
    token  CHAR(43) PRIMARY KEY,
    data   bytea     NOT NULL,
    expiry timestamp NOT NULL
);
--rollback drop table sessions;

--changeset aomar:16
CREATE INDEX sessions_expiry_idx ON sessions (expiry);
--rollback drop index sessions_expiry_idx;