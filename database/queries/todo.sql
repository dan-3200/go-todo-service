-- name: CreateTodo :one
INSERT INTO todo (title, content, status)
VALUES (sqlc.arg(title), sqlc.arg(content), sqlc.arg(status))
RETURNING id;

-- name: GetTodo :one
SELECT id, title, content, status, created_at
FROM todo
WHERE id = $1;

-- name: ListTodos :many
SELECT id, title, content, status, created_at
FROM todo;

-- name: UpdateTodo :exec
UPDATE todo
SET title = sqlc.arg(title),
    content = sqlc.arg(content),
    status = sqlc.arg(status)
WHERE id = sqlc.arg(id);

-- name: DeleteTodo :execrows
DELETE FROM todo
WHERE id = $1;
