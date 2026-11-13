CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL
);

CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    checked BOOLEAN DEFAULT FALSE,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE group_urls (
    group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    url_id INT NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, url_id)
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    url_id INT NOT NULL UNIQUE REFERENCES urls(id) ON DELETE CASCADE,
    seller_name TEXT,
    name TEXT NOT NULL,
    rating NUMERIC(2,1),
    description TEXT,
    options JSONB,
    last_checked TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE price_history (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);