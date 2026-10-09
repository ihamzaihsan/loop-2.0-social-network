DROP TABLE user_ids;
DROP TABLE reports;
DROP TABLE user_blocks;
DROP TABLE password_resets;
DROP TABLE media_uploads;
ALTER TABLE group_events DROP COLUMN creator_id;
ALTER TABLE users DROP COLUMN is_suspended;

ALTER TABLE users DROP COLUMN is_moderator;
