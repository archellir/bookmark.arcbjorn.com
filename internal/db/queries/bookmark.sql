-- name: CreateBookmark :one
INSERT INTO bookmarks (
  name,
  url
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetBookmarkById :one
SELECT * FROM bookmarks
WHERE id = $1 LIMIT 1;

-- name: ListBookmarks :many
SELECT * FROM bookmarks
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateBookmarkName :one
UPDATE bookmarks
SET name = $2
WHERE id = $1
RETURNING *;

-- name: UpdateBookmarkUrl :one
UPDATE bookmarks
SET url = $2
WHERE id = $1
RETURNING *;

-- name: UpdateBookmarkGroupId :one
UPDATE bookmarks
SET group_id = $2
WHERE id = $1
RETURNING *;

-- name: SearchBookmarkByNameAndUrl :many
SELECT * FROM bookmarks  
WHERE
  url ILIKE sqlc.arg(search_string)::text OR
  name ILIKE sqlc.arg(search_string)::text
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: DeleteBookmark :exec
DELETE FROM bookmarks
WHERE id = $1;

-- name: DeleteBookmarks :exec
DELETE FROM bookmarks;