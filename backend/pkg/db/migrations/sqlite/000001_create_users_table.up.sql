CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    dob DATE NOT NULL,
    avatar TEXT,
    nickname TEXT,
    about_me TEXT,
    is_public BOOLEAN,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
