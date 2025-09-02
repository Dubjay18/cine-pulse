-- +goose Up
-- Add rating and source URL columns
ALTER TABLE content ADD COLUMN rating REAL;
ALTER TABLE content ADD COLUMN source_url TEXT;
ALTER TABLE content ADD COLUMN scraped_at DATETIME DEFAULT CURRENT_TIMESTAMP;

-- Create index for rating
CREATE INDEX IF NOT EXISTS idx_content_rating ON content(rating);

-- +goose Down
-- Remove added columns (SQLite doesn't support DROP COLUMN directly)
-- This would require recreating the table in a real scenario
-- For demonstration purposes, we'll just drop the index

DROP INDEX IF EXISTS idx_content_rating;
