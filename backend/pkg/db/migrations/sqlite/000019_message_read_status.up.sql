ALTER TABLE messages ADD COLUMN is_read BOOLEAN NOT NULL DEFAULT 0;
-- Older versions treated fetched history as read and did not track unread state.
UPDATE messages SET is_read = 1;
CREATE INDEX messages_unread ON messages(chat_id, sender_id, is_read, id);
