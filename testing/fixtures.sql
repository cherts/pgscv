-- schema fixtures
CREATE ROLE pgscv WITH LOGIN SUPERUSER;

CREATE DATABASE pgscv_fixtures OWNER pgscv;
\c pgscv_fixtures pgscv

-- create pg_stat_statements
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
SELECT pg_stat_statements_reset();

-- create pgstattuple
CREATE EXTENSION IF NOT EXISTS pgstattuple;

-- create table with invalid index
CREATE TABLE orders (id SERIAL PRIMARY KEY, name TEXT, status INT);
CREATE INDEX orders_status_idx ON orders (status);
UPDATE pg_index SET indisvalid = false WHERE indexrelid = (SELECT oid FROM pg_class WHERE relname = 'orders_status_idx');

-- create table with redundant index
CREATE TABLE products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock_quantity INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE
);
CREATE INDEX products_name_idx ON products (name);
CREATE INDEX products_name_size_idx ON products (name, price);

-- create table with near-to-overflow sequence
CREATE TABLE events (id SERIAL PRIMARY KEY, key TEXT, payload TEXT);
SELECT setval('events_id_seq', 2000000000);

-- create tables with non-indexed foreign key
CREATE TABLE accounts (id SERIAL PRIMARY KEY, name TEXT, email TEXT, passowrd TEXT, status INTEGER);
CREATE TABLE statuses (id SERIAL PRIMARY KEY, name TEXT);
ALTER TABLE accounts ADD CONSTRAINT accounts_status_constraint FOREIGN KEY (status) REFERENCES statuses (id);

-- create tables with foreign key and different columns types
CREATE TABLE persons (id SERIAL PRIMARY KEY, name TEXT, email TEXT, passowrd TEXT, property BIGINT);
CREATE TABLE properties (id SERIAL PRIMARY KEY, name TEXT);
ALTER TABLE persons ADD CONSTRAINT persons_properties_constraint FOREIGN KEY (property) REFERENCES properties (id);

-- create table with no primary/unique key
CREATE TABLE migrations (id INT, created_at TIMESTAMP, description TEXT);
CREATE INDEX migrations_created_at_idx ON migrations (created_at);

-- insert products data
INSERT INTO products (
    name,
    category,
    price,
    stock_quantity,
    description,
    is_active
)
SELECT
    'Product Batch ' || s.id AS name,
    CASE (s.id % 5)
        WHEN 0 THEN 'Electronics'
        WHEN 1 THEN 'Books'
        WHEN 2 THEN 'Home Goods'
        WHEN 3 THEN 'Apparel'
        ELSE 'Miscellaneous'
    END AS category,
    ROUND((RANDOM() * 500 + 10)::numeric, 2) AS price,
    FLOOR(RANDOM() * 200)::int AS stock_quantity,
    'Auto-generated description for product ID ' || s.id || '. Lorem ipsum dolor sit amet, consectetur adipiscing elit.' AS description,
    (s.id % 10 <> 0) AS is_active
FROM generate_series(1, 100) AS s(id);