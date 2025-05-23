CREATE TABLE links (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    original_url TEXT NOT NULL,
    short_url_id TEXT NOT NULL UNIQUE,
    pretty_id TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TRIGGER update_links_updated_at
AFTER UPDATE ON links
FOR EACH ROW
BEGIN
  UPDATE links SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
