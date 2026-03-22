-- Create users table for storing user accounts
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) NOT NULL UNIQUE
);

-- Create index on api_key for efficient lookups
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);
