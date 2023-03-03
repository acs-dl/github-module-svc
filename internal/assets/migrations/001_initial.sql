-- +migrate Up

CREATE TABLE IF NOT EXISTS responses (
    id UUID PRIMARY KEY,
    status TEXT NOT NULL,
    error TEXT
--     payload JSONB,
--     created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    github_id BIGINT PRIMARY KEY,
    id BIGINT UNIQUE,
    username TEXT NOT NULL UNIQUE,
    avatar_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS users_id_idx ON users(id);
CREATE INDEX IF NOT EXISTS users_username_idx ON users(username);
CREATE INDEX IF NOT EXISTS users_githubId_idx ON users(github_id);

CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    link TEXT NOT NULL,
    UNIQUE(link)
);
INSERT INTO links (link) VALUES ('mhrynenko/TESTAPI');
INSERT INTO links (link) VALUES ('acstestapi');

CREATE INDEX IF NOT EXISTS links_link_idx ON links(link);

CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE IF NOT EXISTS subs (
    id BIGINT PRIMARY KEY,
    link TEXT UNIQUE NOT NULL,
    path TEXT NOT NULL,
    type TEXT NOT NULL,
    parent_id BIGINT,
    lpath ltree,

    UNIQUE (id, parent_id)
);

CREATE INDEX IF NOT EXISTS lpath_gist_idx ON subs USING GIST (lpath);
CREATE INDEX IF NOT EXISTS lpath_btree_idx ON subs USING BTREE (lpath);
CREATE INDEX IF NOT EXISTS subs_id_idx ON subs(id);
CREATE INDEX IF NOT EXISTS subs_link_idx ON subs(link);
CREATE INDEX IF NOT EXISTS subs_parentId_idx ON subs(parent_id);

CREATE TABLE IF NOT EXISTS permissions (
    request_id TEXT NOT NULL,
    user_id BIGINT,
    username TEXT NOT NULL,
    github_id INT NOT NULL,
    link TEXT NOT NULL,
    access_level TEXT NOT NULL,
    type TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    has_parent BOOLEAN NOT NULL DEFAULT TRUE,
    has_child BOOLEAN NOT NULL DEFAULT FALSE,

    UNIQUE (github_id, link),
    FOREIGN KEY(github_id) REFERENCES users(github_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY(link) REFERENCES subs(link) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS permissions_userId_idx ON permissions(user_id);
CREATE INDEX IF NOT EXISTS permissions_githubId_idx ON permissions(github_id);
CREATE INDEX IF NOT EXISTS permissions_link_idx ON permissions(link);

-- +migrate Down

DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS responses;
DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS subs;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS users_id_idx;
DROP INDEX IF EXISTS users_username_idx;
DROP INDEX IF EXISTS users_gitlabId_idx;

DROP INDEX IF EXISTS links_link_idx;

DROP INDEX IF EXISTS subs_id_idx;
DROP INDEX IF EXISTS subs_link_idx;
DROP INDEX IF EXISTS subs_parentId_idx;

DROP INDEX IF EXISTS permissions_userId_idx;
DROP INDEX IF EXISTS permissions_gitlabId_idx;
DROP INDEX IF EXISTS permissions_link_idx;

DROP INDEX IF EXISTS lpath_gist_idx;
DROP INDEX IF EXISTS lpath_btree_idx;

DROP EXTENSION IF EXISTS ltree;