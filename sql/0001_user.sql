-- auto-generated definition
-- auto-generated definition
create type gender as enum ('MAN', 'WOMAN', 'OTHER');

alter type gender owner to postgres;


create table users
(
    id           varchar(32) not null
        constraint users_pk
            primary key,
    name         varchar(32) not null,
    password     varchar(32),
    delete_flag  boolean     not null,
    gender       gender      not null,
    phone_number varchar(32)
);

alter table users
    owner to postgres;

