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

CREATE INDEX idx_product_features_quality ON product_features(quality_score DESC);
CREATE INDEX idx_product_features_senitiment ON product_features(senntiment_score DESC);


------------------ Product Embeddings -------------------------------------

CREATE TABLE product_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL UNIQUE REFERENCES products(id) ON DELETE CASCADE,
    embedding VECTOR(384) NOT NULL,
    model TEXT NOT NULL DEFAULT 'all-MiniLM-L6-v2',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_embeddings_hnsw
    ON product_embeddings USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);


------------------- Product Summaries -------------------------------------

CREATE TABLE product_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    porduct_id UUID NOT NULL UNIQUE REFERENCES products(id) ON DELETE CASCADE,
    pros TEXT[] NOT NULL DEFAULT '{}',
    cons TEXT[] NOT NULL DEFAULT '{}',
    summary TEXT,
    verdict TEXT,
    model TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--------------------- Product reviews meta -------------------------------------

CREATE TABLE product_reviews_meta (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL UNIQUE REFERENCES products(id) ON DELETE CASCADE,
    review_count INT NOT NULL DEFAULT 0,
    avg_rating NUMERIC(3, 2),
    rating_distribution JSONB NOT NULL DEFAULT '{}'::jsonb,
    review_velocity NUMERIC(10, 4),
    helpfulness_ratio NUMERIC(5, 4),
    raw_reviews_r2_path TEXT,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

---------------------- Product Users -------------------------------------

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    display_name TEXT,
    avtar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ
);


----------------------- Saved Products -------------------------------------

CREATE TABLE saved_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, product_id)
);

CREATE INDEX idx_saved_products_user_id ON saved_products(user_id, created_at DESC);

----------------------- Saved searches -------------------------------------

CREATE TABLE saved_searches (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query      TEXT NOT NULL,
    filters    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
 
CREATE INDEX idx_saved_searches_user_id ON saved_searches(user_id, created_at DESC);

------------------------ Collections -------------------------------------

CREATE TABLE collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);

CREATE INDEX idx_collections_user_id ON collections(user_id);

------------------------ Collection Products -------------------------------------

CREATE TABLE collection_products (
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_id, product_id)
)

------------------------  Chat session -------------------------------------

CREATE TABLE chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_session_user_id ON chat_sessions(user_id, updated_at DESC);

CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

);

CREATE INDEX idx_chat_messages_session_id ON chat_messages(session_id, created_at ASC);


-------------------------  Cart Items -------------------------------------

CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, product_id)
);