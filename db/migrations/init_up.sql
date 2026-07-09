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

    CONSTRAINT chk_products_published_staus CHECK (published_status IN ('ACTIVE', 'INACTIVE', 'ARCHIVED')),
    CONSTRAINT chk_products_condition CHECK (condition IN ('New', 'Used', 'Refurbished'))
);

CREATE INDEX idx_products_avg_rating ON products(avg_rating DESC NULLS LAST);
CREATE INDEX idx_products_review_count ON products(review_count DESC);
CREATE INDEX idx_products_fts ON products USING GIN(fts_vector);
CREATE INDEX idx_products_brand ON products(brand);
CREATE INDEX idx_products_published ON products(published_status);

CREATE FUNCTION products_fts_trigger() RETURNS trigger AS $$
BEGIN
    NEW.fts_vector := to_tsvector('english',
        coalesce(NEW.title, '') || ' ' ||
        coalesce(NEW.brand, '') || ' ' ||
        coalesce(NEW.description, ''));
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_products_fts
Before INSERT OR UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION products_fts_trigger();


----------------- Product Features -------------------------------------

CREATE TABLE product_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL UNIQUE REFERENCES products(id) ON DELETE CASCADE,
    senetiment_score NUMERIC(5, 4),
    quality_score NUMERIC(5, 4),
    review_velocity NUMERIC(10, 4)
    brand_score NUMERIC(5, 4),
    helpfulness_ration NNUMERIC(5, 4),
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);