-- +migrate Up

create table if not exists responses (
    id uuid primary key,
    status text not null,
    error text,
    payload jsonb,
    created_at timestamp without time zone not null default current_timestamp
);

create table if not exists users (
    github_id bigint primary key,
    id bigint unique,
    username text not null unique,
    avatar_url text not null,
    updated_at timestamp with time zone not null default current_timestamp,
    created_at timestamp with time zone default current_timestamp
);

create index if not exists users_id_idx on users(id);
create index if not exists users_username_idx on users(username);
create index if not exists users_githubid_idx on users(github_id);

create table if not exists links (
    id serial primary key,
    link text not null,
    unique(link)
);
insert into links (link) values ('mhrynenko/testapi');
insert into links (link) values ('acstestapi');
insert into links (link) values ('testapiacsdl');

create index if not exists links_link_idx on links(link);

create table if not exists subs (
    id bigint primary key,
    link text unique not null,
    path text not null,
    type text not null,
    parent_id bigint,

    unique (id, parent_id)
);

create index if not exists subs_id_idx on subs(id);
create index if not exists subs_link_idx on subs(link);
create index if not exists subs_parentid_idx on subs(parent_id);

create table if not exists permissions (
    request_id text not null,
    user_id bigint,
    username text not null,
    github_id int not null,
    link text not null,
    access_level text not null,
    type text not null,
    created_at timestamp without time zone not null,
    expires_at timestamp without time zone not null,
    updated_at timestamp with time zone not null default current_timestamp,
    has_parent boolean not null default true,
    has_child boolean not null default false,
    parent_link text,

    unique (github_id, link),
    foreign key(github_id) references users(github_id) on delete cascade on update cascade,
    foreign key(link) references subs(link) on delete cascade on update cascade
);

create index if not exists permissions_userid_idx on permissions(user_id);
create index if not exists permissions_githubid_idx on permissions(github_id);
create index if not exists permissions_link_idx on permissions(link);

-- +migrate Down

drop index if exists permissions_userid_idx;
drop index if exists permissions_githubid_idx;
drop index if exists permissions_link_idx;

drop table if exists permissions;

drop index if exists subs_id_idx;
drop index if exists subs_link_idx;
drop index if exists subs_parentid_idx;

drop table if exists subs;

drop index if exists links_link_idx;

drop table if exists links;

drop index if exists users_id_idx;
drop index if exists users_username_idx;
drop index if exists users_githubid_idx;

drop table if exists users;
drop table if exists responses;
