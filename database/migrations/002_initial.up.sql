ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified TIMESTAMP;

CREATE TABLE IF NOT EXISTS otp (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    email TEXT,
    phone INT,
    otp_number TEXT NOT NULL,
    created_at TIMESTAMP,
    expires_at TIMESTAMP
);