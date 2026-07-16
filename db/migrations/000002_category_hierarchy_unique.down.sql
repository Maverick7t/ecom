DROP INDEX IF EXISTS idx_categories_parent_slug_unique;
ALTER TABLE categories ADD CONSTRAINT categories_slug_key UNIQUE (slug);