-- name: AddRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    NULL
)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT u.*
FROM refresh_tokens rt
JOIN users u ON u.id = rt.user_id
WHERE rt.token = $1
  AND rt.revoked_at IS NULL
  AND rt.expires_at > NOW();

-- name: RevokeToken :one
UPDATE refresh_tokens 
SET 
    revoked_at = NOW(),
    updated_at = NOW()
WHERE token = $1 AND revoked_at IS NULL
RETURNING *;



