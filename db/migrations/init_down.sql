DROP TABLE IF EXISTS sync_runs;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_sessions;
DROP TABLE IF EXISTS collection_products;
DROP TABLE IF EXISTS collections;
DROP TABLE IF EXISTS saved_searches;
DROP TABLE IF EXISTS saved_products;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS product_reviews_meta;
DROP TABLE IF EXISTS product_summaries;
DROP TABLE IF EXISTS product_embeddings;
DROP TABLE IF EXISTS product_features;
DROP TABLE IF EXISTS product_categories;

DROP TRIGGER IF EXISTS trg_products_fts ON products;
DROP FUNCTION IF EXISTS products_fts_trigger();

DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;