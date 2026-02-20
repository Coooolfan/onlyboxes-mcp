-- name: ListTrustedTokensByAccount :many
SELECT
    token_id,
    account_id,
    name,
    name_key,
    token_hash,
    token_masked,
    generated,
    created_at_unix_ms,
    updated_at_unix_ms
FROM trusted_tokens
WHERE account_id = ?
ORDER BY created_at_unix_ms ASC, token_id ASC;

-- name: GetTrustedTokenByID :one
SELECT
    token_id,
    account_id,
    name,
    name_key,
    token_hash,
    token_masked,
    generated,
    created_at_unix_ms,
    updated_at_unix_ms
FROM trusted_tokens
WHERE token_id = ?
LIMIT 1;

-- name: GetTrustedTokenByAccountAndNameKey :one
SELECT
    token_id,
    account_id,
    name,
    name_key,
    token_hash,
    token_masked,
    generated,
    created_at_unix_ms,
    updated_at_unix_ms
FROM trusted_tokens
WHERE account_id = ? AND name_key = ?
LIMIT 1;

-- name: GetTrustedTokenByHash :one
SELECT
    token_id,
    account_id,
    name,
    name_key,
    token_hash,
    token_masked,
    generated,
    created_at_unix_ms,
    updated_at_unix_ms
FROM trusted_tokens
WHERE token_hash = ?
LIMIT 1;

-- name: InsertTrustedToken :exec
INSERT INTO trusted_tokens (
    token_id,
    account_id,
    name,
    name_key,
    token_hash,
    token_masked,
    generated,
    created_at_unix_ms,
    updated_at_unix_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteTrustedTokenByIDAndAccount :execrows
DELETE FROM trusted_tokens
WHERE token_id = ? AND account_id = ?;
