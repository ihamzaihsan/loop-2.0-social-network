DROP INDEX IF EXISTS messages_unread;
ALTER TABLE messages DROP COLUMN is_read;
