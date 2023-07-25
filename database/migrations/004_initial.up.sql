ALTER TABLE users
DROP COLUMN is_verified,
ADD COLUMN email_verified TIMESTAMP DEFAULT NULL,
ADD COLUMN phone_verified TIMESTAMP DEFAULT NULL;

DROP TABLE otp;

CREATE TABLE IF NOT EXISTS otp (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    user_id UUID REFERENCES users(user_id),
    typee TEXT,
    created_at TIMESTAMP,
    expires_at TIMESTAMP
);