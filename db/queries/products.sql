-- name: GetProductByID :one
SELECT *
FROM products
WHERE id = $1;