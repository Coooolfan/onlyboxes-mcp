-- name: CountAdminAccounts :one
SELECT COUNT(*)
FROM accounts
WHERE is_admin = 1;

-- name: GetAccountByID :one
SELECT
    account_id,
    username,
    username_key,
    password_hash,
    hash_algo,
    is_admin,
    created_at_unix_ms,
    updated_at_unix_ms
FROM accounts
WHERE account_id = ?
LIMIT 1;

-- name: GetAccountByUsernameKey :one
SELECT
    account_id,
    username,
    username_key,
    password_hash,
    hash_algo,
    is_admin,
    created_at_unix_ms,
    updated_at_unix_ms
FROM accounts
WHERE username_key = ?
LIMIT 1;

-- name: GetFirstAdminAccount :one
SELECT
    account_id,
    username,
    username_key,
    password_hash,
    hash_algo,
    is_admin,
    created_at_unix_ms,
    updated_at_unix_ms
FROM accounts
WHERE is_admin = 1
ORDER BY created_at_unix_ms ASC, account_id ASC
LIMIT 1;

-- name: InsertAccount :exec
INSERT INTO accounts (
    account_id,
    username,
    username_key,
    password_hash,
    hash_algo,
    is_admin,
    created_at_unix_ms,
    updated_at_unix_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);
