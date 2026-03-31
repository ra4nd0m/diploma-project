-- +goose Up
create table access_mode (
    id bigserial primary key,
    code text not null unique,
    name text not null
);

create table issuance_kind (
    id bigserial primary key,
    code text not null unique,
    name text not null
);

create table condition_type (
    id bigserial primary key,
    code text not null unique,
    name text not null
);

create table achievement_status (
    id bigserial primary key,
    code text not null unique,
    name text not null
);

create table achievement (
    id bigserial primary key,
    name text not null,
    description text not null,
    icon_link text null,
    owner_id uuid not null,
    cohort_id bigint not null,
    access_mode bigint not null references access_mode(id),
    issuance_kind bigint not null references issuance_kind(id),
    condition_type bigint null references condition_type(id),
    condition_payload jsonb null,
    created_on timestamptz not null default now(),
    constraint chk_achievement_condition_pair check (
        (
            condition_type is null
            and condition_payload is null
        )
        or (
            condition_type is not null
            and condition_payload is not null
        )
    ),
    constraint chk_achievement_condition_payload_object check (
        condition_payload is null
        or jsonb_typeof(condition_payload) = 'object'
    )
);

create table achievement_issuance (
    id bigserial primary key,
    achievement_id bigint not null references achievement(id) on delete cascade,
    recipient_id uuid not null,
    issuer_id uuid not null,
    status bigint not null references achievement_status(id),
    additional_detail text null,
    progress_payload jsonb null,
    created_on timestamptz not null default now(),
    constraint chk_achievement_issuance_progress_payload_object check (
        progress_payload is null
        or jsonb_typeof(progress_payload) = 'object'
    )
);

create unique index ux_achievement_issuance_achievement_recipient on achievement_issuance (achievement_id, recipient_id);

create index ix_achievement_cohort_id on achievement (cohort_id);

create index ix_achievement_owner_id on achievement (owner_id);

create index ix_achievement_access_mode on achievement (access_mode);

create index ix_achievement_issuance_kind on achievement (issuance_kind);

create index ix_achievement_condition_type on achievement (condition_type);

create index ix_achievement_condition_payload_gin on achievement using gin (condition_payload);

create index ix_achievement_issuance_recipient_id on achievement_issuance (recipient_id);

create index ix_achievement_issuance_issuer_id on achievement_issuance (issuer_id);

create index ix_achievement_issuance_status on achievement_issuance (status);

create index ix_achievement_issuance_progress_payload_gin on achievement_issuance using gin (progress_payload);

insert into
    access_mode (code, name)
values
    ('cohort', 'Cohort'),
    ('private', 'Private'),
    ('public', 'Public');

insert into
    issuance_kind (code, name)
values
    ('manual', 'Manual'),
    ('automatic', 'Automatic');

insert into
    condition_type (code, name)
values
    ('all_of', 'All Of'),
    ('x_of_m', 'X Of M');

insert into
    achievement_status (code, name)
values
    ('issued', 'Issued'),
    ('in_progress', 'In Progress'),
    ('closed', 'Closed'),
    ('revoked', 'Revoked');

-- +goose Down
drop index if exists ix_achievement_issuance_progress_payload_gin;

drop index if exists ix_achievement_issuance_status;

drop index if exists ix_achievement_issuance_issuer_id;

drop index if exists ix_achievement_issuance_recipient_id;

drop index if exists ix_achievement_condition_payload_gin;

drop index if exists ix_achievement_condition_type;

drop index if exists ix_achievement_issuance_kind;

drop index if exists ix_achievement_access_mode;

drop index if exists ix_achievement_owner_id;

drop index if exists ix_achievement_cohort_id;

drop index if exists ux_achievement_issuance_achievement_recipient;

drop table if exists achievement_issuance;

drop table if exists achievement;

drop table if exists achievement_status;

drop table if exists condition_type;

drop table if exists issuance_kind;

drop table if exists access_mode;