package query

import "database/sql"

func DeleteGroupTx(tx *sql.Tx, id int) error {
	statements := []string{
		`DELETE FROM event_responses WHERE event_id IN(SELECT id FROM group_events WHERE group_id=?)`,
		`DELETE FROM event_response_options WHERE event_id IN(SELECT id FROM group_events WHERE group_id=?)`,
		`DELETE FROM group_events WHERE group_id=?`,
		`DELETE FROM group_comments WHERE post_id IN(SELECT id FROM group_posts WHERE group_id=?)`,
		`DELETE FROM group_posts WHERE group_id=?`,
		`DELETE FROM group_chat_messages WHERE group_id=?`,
		`DELETE FROM group_invitations WHERE group_id=?`,
		`DELETE FROM group_join_requests WHERE group_id=?`,
		`DELETE FROM group_members WHERE group_id=?`,
		`DELETE FROM notifications WHERE related_id=? AND type IN('group_invitation','group_join_request','group_invitation_accepted','group_join_request_accepted','group_event')`,
		`DELETE FROM groups WHERE id=?`,
	}
	for _, s := range statements {
		if _, err := tx.Exec(s, id); err != nil {
			return err
		}
	}
	return nil
}

func DeleteContentTx(tx *sql.Tx, kind string, id int) error {
	var statements []string
	switch kind {
	case "post":
		statements = []string{`DELETE FROM comments WHERE post_id=?`, `DELETE FROM likes WHERE post_id=?`, `DELETE FROM post_viewers WHERE post_id=?`, `DELETE FROM posts WHERE id=?`}
	case "comment":
		statements = []string{`DELETE FROM comments WHERE id=?`}
	case "group_post":
		statements = []string{`DELETE FROM group_comments WHERE post_id=?`, `DELETE FROM group_posts WHERE id=?`}
	case "group_comment":
		statements = []string{`DELETE FROM group_comments WHERE id=?`}
	case "group_event":
		statements = []string{`DELETE FROM event_responses WHERE event_id=?`, `DELETE FROM event_response_options WHERE event_id=?`, `DELETE FROM group_events WHERE id=?`}
	default:
		return sql.ErrNoRows
	}
	for _, s := range statements {
		if _, err := tx.Exec(s, id); err != nil {
			return err
		}
	}
	return nil
}

func DeleteAccountTx(tx *sql.Tx, id int) error {
	rows, err := tx.Query(`SELECT id FROM groups WHERE creator_id=?`, id)
	if err != nil {
		return err
	}
	var groups []int
	for rows.Next() {
		var group int
		if err = rows.Scan(&group); err != nil {
			rows.Close()
			return err
		}
		groups = append(groups, group)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, group := range groups {
		if err = DeleteGroupTx(tx, group); err != nil {
			return err
		}
	}
	statements := []string{
		`DELETE FROM group_comments WHERE user_id=? OR post_id IN(SELECT id FROM group_posts WHERE user_id=?)`,
		`DELETE FROM group_posts WHERE user_id=?`, `DELETE FROM group_chat_messages WHERE user_id=?`,
		`DELETE FROM group_invitations WHERE inviter_id=? OR invitee_id=?`,
		`DELETE FROM group_join_requests WHERE user_id=?`, `DELETE FROM group_members WHERE user_id=?`,
		`DELETE FROM event_responses WHERE user_id=?`,
		`DELETE FROM comments WHERE user_id=? OR post_id IN(SELECT id FROM posts WHERE user_id=?)`,
		`DELETE FROM likes WHERE user_id=? OR post_id IN(SELECT id FROM posts WHERE user_id=?)`,
		`DELETE FROM post_viewers WHERE viewer_id=? OR post_id IN(SELECT id FROM posts WHERE user_id=?)`,
		`DELETE FROM posts WHERE user_id=?`,
		`DELETE FROM messages WHERE sender_id=? OR chat_id IN(SELECT id FROM chats WHERE user1_id=? OR user2_id=?)`,
		`DELETE FROM chats WHERE user1_id=? OR user2_id=?`,
		`DELETE FROM notifications WHERE user_id=? OR from_user_id=?`,
		`DELETE FROM followers WHERE follower_id=? OR following_id=?`,
		`DELETE FROM users WHERE id=?`,
	}
	for _, s := range statements {
		args := make([]interface{}, 0)
		for _, c := range s {
			if c == '?' {
				args = append(args, id)
			}
		}
		if _, err = tx.Exec(s, args...); err != nil {
			return err
		}
	}
	return nil
}
