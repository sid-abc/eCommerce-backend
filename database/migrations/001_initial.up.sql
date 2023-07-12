CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    number INT,
    address TEXT,
    zip_code INT,
    country TEXT,
    archived TIMESTAMP
);

CREATE TABLE uploads (
    upload_id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    path TEXT NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL
);

CREATE TABLE items (
    item_id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    features TEXT,
    price INT,
    type TEXT NOT NULL,
    stock_no INT NOT NULL,
    archived TIMESTAMP
);

CREATE TABLE carts (
    cart_id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    cart_name TEXT,
    user_id UUID REFERENCES users(user_id)
);

CREATE TABLE images (
    item_id UUID REFERENCES items(item_id) NOT NULL,
    upload_id UUID REFERENCES uploads(upload_id)
);

CREATE TABLE cart_items (
    cart_id UUID REFERENCES carts(cart_id) NOT NULL,
    item_id UUID REFERENCES items(item_id) NOT NULL,
    quantity INT NOT NULL
);

CREATE TABLE user_role (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    role_user TEXT NOT NULL,
    user_id UUID REFERENCES users(user_id)
);

INSERT INTO users(name, email, password) VALUES ('Siddhant', 'abc@gmail.com', '$2a$12$iwTamkdHO27Rhc43SfV2.u0AYT096KDAZOiSyX9FrOaU1h5NSGR8i');

WITH siddhant_user AS (
    SELECT user_id
    FROM users
    WHERE name = 'Siddhant'
    LIMIT 1
    )
INSERT INTO user_role (user_id, role_user)
SELECT user_id, 'admin'
FROM siddhant_user;