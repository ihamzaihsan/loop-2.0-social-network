// Package access centralizes audience and blocking rules for HTTP and socket writes.
package access

import (
	"fmt"
	"socialNetwork/pkg/db"
)

func Blocked(a, b int) bool {
	var blocked bool
	err := db.DBInstance.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM user_blocks WHERE (blocker_id=? AND blocked_id=?) OR (blocker_id=? AND blocked_id=?))`, a, b, b, a).Scan(&blocked)
	return err != nil || blocked
}

// PostAudience is used only with internal integer IDs and known SQL aliases.
func PostAudience(viewer int) string {
	v := fmt.Sprint(viewer)
	return `(p.user_id=` + v + ` OR (u.is_suspended=0
 AND NOT EXISTS(SELECT 1 FROM user_blocks b WHERE (b.blocker_id=` + v + ` AND b.blocked_id=p.user_id) OR (b.blocker_id=p.user_id AND b.blocked_id=` + v + `))
 AND (COALESCE(u.isprivate,0)=0 OR EXISTS(SELECT 1 FROM followers f WHERE f.follower_id=` + v + ` AND f.following_id=p.user_id AND f.status='accept'))
 AND (p.privacy='public' OR (p.privacy='friends' AND EXISTS(SELECT 1 FROM followers f JOIN followers r ON r.follower_id=f.following_id AND r.following_id=f.follower_id WHERE f.follower_id=` + v + ` AND f.following_id=p.user_id AND f.status='accept' AND r.status='accept'))
 OR (p.privacy='private' AND EXISTS(SELECT 1 FROM post_viewers pv WHERE pv.post_id=p.id AND pv.viewer_id=` + v + `)))))`
}

func CanViewPost(viewer, post int) bool {
	var allowed bool
	err := db.DBInstance.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM posts p JOIN users u ON u.id=p.user_id WHERE p.id=? AND `+PostAudience(viewer)+`)`, post).Scan(&allowed)
	return err == nil && allowed
}

func GroupMember(group, user int) bool {
	var allowed bool
	err := db.DBInstance.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id=? AND user_id=? AND status='active')`, group, user).Scan(&allowed)
	return err == nil && allowed
}

func GroupCreator(group, user int) bool {
	var allowed bool
	err := db.DBInstance.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM groups WHERE id=? AND creator_id=?)`, group, user).Scan(&allowed)
	return err == nil && allowed
}
