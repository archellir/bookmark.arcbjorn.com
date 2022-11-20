-- name: CreateUser :one
INSERT INTO users (
  username,
  hashed_password
) VALUES (
  $1, $2
) RETURNING id, username, created_at;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUserPassword :one
UPDATE users
SET hashed_password = $2
WHERE username = $1
RETURNING id, username, created_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE username = $1;