-- +migrate Up

CREATE TABLE IF NOT EXISTS responses (
    id UUID PRIMARY KEY,
    status TEXT NOT NULL,
    error TEXT
--     payload JSONB,
--     created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNIQUE,
    username TEXT NOT NULL UNIQUE,
    github_id BIGINT PRIMARY KEY
);

CREATE INDEX IF NOT EXISTS users_idx ON users(id, username, github_id);

CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    link TEXT NOT NULL,
    UNIQUE(link)
);
INSERT INTO links (link) VALUES ('mhrynenko/TESTAPI');
INSERT INTO links (link) VALUES ('Genobank/BioMessenger');

CREATE INDEX IF NOT EXISTS links_link_idx ON links(link);

CREATE TABLE IF NOT EXISTS permissions (
    request_id TEXT NOT NULL,
    user_id BIGINT,
    username TEXT NOT NULL,
    github_id INT NOT NULL,
    link TEXT NOT NULL,
    permission TEXT NOT NULL,
    type_to TEXT NOT NULL,

    UNIQUE (github_id, link),
    FOREIGN KEY(github_id) REFERENCES users(github_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS permissions_idx ON permissions(user_id, github_id, link);

-- +migrate Down

DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS requests;
DROP TABLE IF EXISTS responses;
DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS verified_users;
DROP TABLE IF EXISTS users;
