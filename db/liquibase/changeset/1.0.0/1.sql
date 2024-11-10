--liquibase formatted sql
--changeset aomar:1
create table mezanas
(
  id uuid not null PRIMARY KEY,
  name varchar(500) not null,
  created_at timestamp not null
);
--rollback drop table mezanas;
