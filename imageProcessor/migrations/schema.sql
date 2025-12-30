CREATE TABLE images (
    id TEXT PRIMARY KEY,
    original_filename TEXT,
    original_path TEXT,
    mime_type TEXT,
    file_size INTEGER,
    status TEXT DEFAULT 'pending',
    processed_versions TEXT, -- JSON: [{"type":"resize","path":"..."}]
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);