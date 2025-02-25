CREATE TABLE group_events (
    id INTEGER PRIMARY KEY,
    group_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    event_time DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id)
);
