CREATE TABLE link_analytics (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    country_code TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (link_id) REFERENCES links(id) ON DELETE CASCADE
);
