-- name: UpsertProduct :one
INSERT INTO products (
    parent_asin, title, brand, description, price, currency,
    image_url, product_type, condition
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (parent_asin) DO UPDATE SET
    title        = EXCLUDED.title,
    brand        = EXCLUDED.brand,
    description  = EXCLUDED.description,
    price        = EXCLUDED.price,
    currency     = EXCLUDED.currency,
    image_url    = EXCLUDED.image_url,
    product_type = EXCLUDED.product_type,
    condition    = EXCLUDED.condition,
    updated_at   = NOW()
RETURNING id, parent_asin;

-- name: LinkProductCategory :exec
INSERT INTO product_categories (product_id, category_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetOrCreateCategory :one
INSERT INTO categories (slug, name)
VALUES ($1, $2)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- name: CreateSyncRun :one
INSERT INTO sync_runs (job_type, status)
VALUES ($1, 'RUNNING')
RETURNING id;

-- name: UpdateSyncRunProgress :exec
UPDATE sync_runs
SET records_in = $2, records_out = $3
WHERE id = $1;

-- name: CompleteSyncRun :exec
UPDATE sync_runs
SET status = $2, records_in = $3, records_out = $4,
    error_message = $5, completed_at = NOW()
WHERE id = $1;