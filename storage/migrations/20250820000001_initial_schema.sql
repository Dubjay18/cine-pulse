-- +goose Up
-- Create the initial content table
CREATE TABLE IF NOT EXISTS content (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    year INTEGER,
    category TEXT NOT NULL,
    extra_info TEXT,
    type TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_content_title ON content(title);
CREATE INDEX IF NOT EXISTS idx_content_type ON content(type);
CREATE INDEX IF NOT EXISTS idx_content_category ON content(category);
CREATE INDEX IF NOT EXISTS idx_content_year ON content(year);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_content_year;
DROP INDEX IF EXISTS idx_content_category;
DROP INDEX IF EXISTS idx_content_type;
DROP INDEX IF EXISTS idx_content_title;

-- Drop table
DROP TABLE IF EXISTS content;
