CREATE TABLE event_response_options (
    id INTEGER PRIMARY KEY,
    event_id INTEGER NOT NULL,
    option_text TEXT NOT NULL,
    FOREIGN KEY (event_id) REFERENCES group_events(id)
);

CREATE TABLE event_responses (
    id INTEGER PRIMARY KEY,
    event_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    response_option_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES group_events(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (response_option_id) REFERENCES event_response_options(id)
);
