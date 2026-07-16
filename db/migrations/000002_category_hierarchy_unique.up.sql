ALTER TABLE categories DROP CONSTRAINT categories_slug_key;

-- NULLS NOT DISTINCT requires Postgres 15+. Verify your Supabase instance
-- version before applying — if older, use a partial unique index instead
-- (one for parent_id IS NULL, one for parent_id IS NOT NULL).
CREATE UNIQUE INDEX idx_categories_parent_slug_unique
    ON categories (parent_id, slug) NULLS NOT DISTINCT;