ALTER TABLE users ADD COLUMN is_suspended BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN is_moderator BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE group_events ADD COLUMN creator_id INTEGER REFERENCES users(id) ON DELETE SET NULL;
UPDATE group_events SET creator_id=(SELECT creator_id FROM groups WHERE groups.id=group_events.group_id);
CREATE TABLE media_uploads(path TEXT PRIMARY KEY, owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE);
CREATE TABLE password_resets (
 token_hash TEXT PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 expires_at DATETIME NOT NULL
);
CREATE TABLE user_blocks (
 blocker_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 blocked_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY(blocker_id, blocked_id),
 CHECK(blocker_id != blocked_id)
);
CREATE TABLE reports (
 id INTEGER PRIMARY KEY,
 reporter_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
 target_type TEXT NOT NULL CHECK(target_type IN ('user','post','comment','group_post','group_comment','group_event')),
 target_id INTEGER NOT NULL,
 reason TEXT NOT NULL,
 status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','resolved','dismissed')),
 resolution TEXT NOT NULL DEFAULT '',
 resolved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 resolved_at DATETIME
);
CREATE INDEX reports_status ON reports(status, id);
CREATE INDEX user_blocks_target ON user_blocks(blocked_id, blocker_id);

-- Keep account identities unique even after deletion, including configured moderators.
CREATE TABLE user_ids(id INTEGER PRIMARY KEY AUTOINCREMENT);
INSERT INTO user_ids(id) SELECT id FROM users;
