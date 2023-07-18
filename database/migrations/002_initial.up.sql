ALTER TABLE users ADD COLUMN is_verified TIMESTAMP;

CREATE TABLE otp (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    email TEXT REFERENCES users(email),
    phone INT REFERENCES users(number),
    otp_number INT NOT NULL,
    created_at TIMESTAMP,
    expires_at TIMESTAMP
)