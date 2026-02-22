-- Create links table for storing shortened URLs
CREATE TABLE IF NOT EXISTS links (
    code VARCHAR(255) PRIMARY KEY,
    url TEXT NOT NULL,
    owner_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on owner_id for efficient user queries
CREATE INDEX IF NOT EXISTS idx_links_owner_id ON links(owner_id);
