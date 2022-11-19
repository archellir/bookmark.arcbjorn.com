-- name: CreateGroup :one
INSERT INTO groups (
  name
) VALUES (
  $1
) RETURNING *;

-- name: GetGroupById :one
SELECT * FROM groups
WHERE id = $1 LIMIT 1;

-- name: ListGroups :many
SELECT * FROM groups
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateGroupName :one
UPDATE groups
SET name = $2
WHERE id = $1
RETURNING *;

-- name: SearchGroupByName :many
SELECT * FROM groups  
WHERE
  name ILIKE sqlc.arg(search_string)::text;

-- name: DeleteGroup :exec
DELETE FROM groups
WHERE id = $1;

-- name: DeleteGroups :exec
DELETE FROM groups;