-- name: CreateBookmark :one
INSERT INTO bookmarks (
  name,
  search_tokens,
  url
) VALUES (
  $1, to_tsvector($1), $2
) RETURNING *;

-- name: GetBookmarkById :one
SELECT * FROM bookmarks
WHERE id = $1 LIMIT 1;

-- name: GetBookmark :one
SELECT * FROM bookmarks
WHERE id = $1 LIMIT 1;

-- name: ListBookmarks :many
SELECT * FROM bookmarks
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateBookmarkName :one
UPDATE bookmarks
SET
  name = $2,
  search_tokens = to_tsvector($2)
WHERE id = $1
RETURNING *;

-- name: UpdateBookmarkUrl :one
UPDATE bookmarks
SET url = $2
WHERE id = $1
RETURNING *;

-- name: SearchBookmarkByNameAndUrl :many
SELECT * FROM bookmarks  
WHERE
  url LIKE '%' + sqlc.arg(search_string)::text + '%' OR
  name @@ to_tsquery(sqlc.arg(search_string)::text);

-- name: DeleteBookmark :exec
DELETE FROM bookmarks
WHERE id = $1;

-- name: DeleteBookmarks :exec
DELETE FROM bookmarks;