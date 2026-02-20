-- +goose Up
CREATE TABLE accounts (
    account_id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    username_key TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    hash_algo TEXT NOT NULL,
    is_admin INTEGER NOT NULL CHECK (is_admin IN (0, 1)),
    created_at_unix_ms INTEGER NOT NULL,
    updated_at_unix_ms INTEGER NOT NULL,
    UNIQUE (username_key)
);

CREATE INDEX idx_accounts_created
    ON accounts(created_at_unix_ms);

INSERT INTO accounts (
    account_id,
    username,
    username_key,
    password_hash,
    hash_algo,
    is_admin,
    created_at_unix_ms,
    updated_at_unix_ms
)
SELECT
    'acc_' || lower(hex(randomblob(16))) AS account_id,
    username,
    lower(trim(username)) AS username_key,
    password_hash,
    hash_algo,
    1 AS is_admin,
    created_at_unix_ms,
    updated_at_unix_ms
FROM dashboard_credentials
WHERE singleton_id = 1;

CREATE TABLE trusted_tokens_new (
    token_id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    token_masked TEXT NOT NULL,
    generated INTEGER NOT NULL CHECK (generated IN (0, 1)),
    created_at_unix_ms INTEGER NOT NULL,
    updated_at_unix_ms INTEGER NOT NULL,
    UNIQUE (account_id, name_key),
    UNIQUE (token_hash),
    FOREIGN KEY (account_id) REFERENCES accounts(account_id) ON DELETE CASCADE
);

INSERT INTO trusted_tokens_new (
    token_id,
    account_id,
    name,
    name_key,
    token_hash,
    token_masked,
    generated,
    created_at_unix_ms,
    updated_at_unix_ms
)
SELECT
    token_id,
    (SELECT account_id
     FROM accounts
     WHERE is_admin = 1
     ORDER BY created_at_unix_ms ASC, account_id ASC
     LIMIT 1) AS account_id,
    name,
    name_key,
    token_hash,
    token_masked,
    generated,
    created_at_unix_ms,
    updated_at_unix_ms
FROM trusted_tokens;

DROP TABLE trusted_tokens;

ALTER TABLE trusted_tokens_new RENAME TO trusted_tokens;

CREATE INDEX idx_trusted_tokens_account_created
    ON trusted_tokens(account_id, created_at_unix_ms);

UPDATE tasks
SET owner_id = (
    SELECT account_id
    FROM trusted_tokens
    WHERE trusted_tokens.token_hash = tasks.owner_id
    LIMIT 1
)
WHERE EXISTS (
    SELECT 1
    FROM trusted_tokens
    WHERE trusted_tokens.token_hash = tasks.owner_id
);

-- +goose Down
CREATE TABLE trusted_tokens_old (
    token_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    token_masked TEXT NOT NULL,
    generated INTEGER NOT NULL CHECK (generated IN (0, 1)),
    created_at_unix_ms INTEGER NOT NULL,
    updated_at_unix_ms INTEGER NOT NULL,
    UNIQUE (name_key),
    UNIQUE (token_hash)
);

INSERT INTO trusted_tokens_old (
    token_id,
    name,
    name_key,
    token_hash,
    token_masked,
    generated,
    created_at_unix_ms,
    updated_at_unix_ms
)
SELECT
    token_id,
    name,
    name_key,
    token_hash,
    token_masked,
    generated,
    created_at_unix_ms,
    updated_at_unix_ms
FROM trusted_tokens;

DROP TABLE trusted_tokens;

ALTER TABLE trusted_tokens_old RENAME TO trusted_tokens;

CREATE INDEX idx_trusted_tokens_created
    ON trusted_tokens(created_at_unix_ms);

DROP INDEX IF EXISTS idx_trusted_tokens_account_created;
DROP INDEX IF EXISTS idx_accounts_created;
DROP TABLE IF EXISTS accounts;
