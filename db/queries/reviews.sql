-- NOT currently used by reviews_ingestion.go — that worker uses raw
-- pgx queries directly (see MIGRATION_GUIDE.md for why: avoiding
-- hand-fabricated sqlc output that might drift from your actual
-- sqlc v1.31.1 formatting). Run `sqlc generate` after adding this file
-- if you want reviews.Worker rewired onto dbgen for consistency with
-- catalog.Worker's pattern — that's a follow-up, not required to ship.

-- name: UpsertProductReviewsMeta :exec
INSERT INTO product_reviews_meta
    (product_id, review_count, avg_rating, rating_distribution,
     review_velocity, helpfulness_ratio, raw_reviews_storage_path)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (product_id) DO UPDATE SET
    review_count             = EXCLUDED.review_count,
    avg_rating                = EXCLUDED.avg_rating,
    rating_distribution        = EXCLUDED.rating_distribution,
    review_velocity            = EXCLUDED.review_velocity,
    helpfulness_ratio          = EXCLUDED.helpfulness_ratio,
    raw_reviews_storage_path   = EXCLUDED.raw_reviews_storage_path,
    computed_at                = NOW();

-- name: UpdateProductAggregatesFromReviews :exec
UPDATE products SET avg_rating = $1, review_count = $2, updated_at = NOW()
WHERE id = $3;

-- name: GetKnownProductsByCategory :many
SELECT p.parent_asin, p.id
FROM products p
JOIN product_categories pc ON pc.product_id = p.id
JOIN categories c ON c.id = pc.category_id
WHERE c.slug = $1 OR c.name = $1;
