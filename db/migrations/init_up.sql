---------------- Extensions -------------------------------------
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS vector;


---------------- Categories -------------------------------------
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    product_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);


----------------- Products -------------------------------------
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_asin TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    brand TEXT,
    description TEXT,
    price NUMERIC(10, 2),
    currency TEXT NOT NULL DEFAULT 'USD',
    image_url TEXT,
    product_type TEXT,
    condition TEXT NOT NULL DEFAULT 'New',
    avg_rating NUMERIC(3, 2),
    review_count INT NOT NULL DEFAULT 0,
    rating_distribution JSONB NOT NULL DEFAULT '{}'::jsonb,
    published_status TEXT NOT NULL DEFAULT 'ACTIVE',
    content_hash TEXT,
    fts_vector TSVECTOR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
)