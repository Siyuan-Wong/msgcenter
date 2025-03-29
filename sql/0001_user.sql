create type gender as enum ('M', 'F', 'O');

alter type gender owner to postgres;

create table users
(
    id           varchar(32)       not null
        constraint users_pk
            primary key,
    name         varchar(32)       not null,
    password     varchar(32),
    delete_flag  boolean           not null,
    gender       gender            not null,
    phone_number varchar(32),
    age          integer default 0 not null,
    email        varchar(64)
);

alter table users
    owner to postgres;

create index users_name_index
    on users (name);

create index users_email_index
    on users (email);

create index users_phone_number_index
    on users (phone_number);

create type device_type as enum ('Mobile', 'Desktop', 'Other');

alter type device_type owner to postgres;

create table devices
(
    id               varchar(32)                              not null
        constraint devices_pk_2
            primary key,
    client_device_id varchar(32)                              not null
        constraint devices_pk
            unique,
    last_active_time timestamp   default CURRENT_TIMESTAMP    not null,
    actived          boolean     default true                 not null,
    delete_flag      boolean     default false                not null,
    curr_user_id     varchar(32)                              not null
        constraint devices_users_id_fk
            references users
            on delete set null
            deferrable,
    device_type      device_type default 'Other'::device_type not null
);

alter table devices
    owner to postgres;

create table user_device_relations
(
    id          varchar(32) not null
        constraint user_device_relations_pk
            primary key,
    user_id     varchar(32) not null
        constraint user_device_relations_users_id_fk
            references users
            on delete cascade
            deferrable,
    device_id   varchar(32) not null
        constraint user_device_relations_devices_id_fk
            references devices
            on delete restrict
            deferrable,
    tag         varchar(32)
);

alter table user_device_relations
    owner to postgres;

