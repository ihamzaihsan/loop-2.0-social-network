CREATE TABLE oauth_identities (
 provider TEXT NOT NULL CHECK(provider='google'),
 subject TEXT NOT NULL,
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 PRIMARY KEY(provider, subject),
 UNIQUE(provider, user_id)
);
CREATE TABLE oauth_flows (
 token_hash TEXT PRIMARY KEY,
 verifier TEXT NOT NULL DEFAULT '',
 intent TEXT NOT NULL CHECK(intent IN ('signin','link','reauth','register')),
 user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
 subject TEXT NOT NULL DEFAULT '',
 email TEXT NOT NULL DEFAULT '',
 first_name TEXT NOT NULL DEFAULT '',
 last_name TEXT NOT NULL DEFAULT '',
 expires_at INTEGER NOT NULL
);
CREATE INDEX oauth_flows_expiry ON oauth_flows(expires_at);
ALTER TABLE sessions ADD COLUMN auth_provider TEXT NOT NULL DEFAULT 'password';
