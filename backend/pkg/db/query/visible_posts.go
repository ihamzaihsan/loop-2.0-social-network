package query

import (
	"database/sql"
	"socialNetwork/pkg/access"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func VisiblePosts(viewer, target, limit, offset int, search string) ([]models.PostResponse, int, error) {
	condition := access.PostAudience(viewer) + ` AND (?=0 OR p.user_id=?) AND (?='' OR instr(lower(COALESCE(p.content,'')),lower(?))>0)`
	var total int
	if err := db.DBInstance.DB.QueryRow(`SELECT COUNT(*) FROM posts p JOIN users u ON u.id=p.user_id WHERE `+condition, target, target, search, search).Scan(&total); err != nil {
		return nil, 0, err
	}
	statement := `SELECT p.id,p.user_id,COALESCE(p.content,''),COALESCE(p.image,''),p.privacy,p.created_at,
 u.first_name,u.last_name,u.nickname,u.avatar,
 (SELECT COUNT(*) FROM likes l WHERE l.post_id=p.id),EXISTS(SELECT 1 FROM likes l WHERE l.post_id=p.id AND l.user_id=?)
 FROM posts p JOIN users u ON u.id=p.user_id WHERE ` + condition + ` ORDER BY p.created_at DESC,p.id DESC `
	args := []any{viewer, target, target, search, search}
	if limit >= 0 {
		statement += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}
	rows, err := db.DBInstance.DB.Query(statement, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	posts := make([]models.PostResponse, 0)
	for rows.Next() {
		var p models.PostResponse
		var nickname, avatar sql.NullString
		if err = rows.Scan(&p.ID, &p.UserID, &p.Content, &p.Image, &p.Privacy, &p.CreatedAt, &p.Author.FirstName, &p.Author.LastName, &nickname, &avatar, &p.LikeCount, &p.IsLiked); err != nil {
			return nil, 0, err
		}
		p.Author.ID = p.UserID
		p.Image = fixImagePath(p.Image)
		if nickname.Valid {
			p.Author.Nickname = &nickname.String
		}
		if avatar.Valid {
			value := fixImagePath(avatar.String)
			p.Author.Avatar = &value
		}
		posts = append(posts, p)
	}
	return posts, total, rows.Err()
}
