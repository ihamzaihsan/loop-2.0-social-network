UPDATE sessions SET is_active=0 WHERE is_active=1 AND rowid NOT IN (SELECT MAX(rowid) FROM sessions WHERE is_active=1 GROUP BY user_id);
CREATE UNIQUE INDEX sessions_one_active_user ON sessions(user_id) WHERE is_active=1;
