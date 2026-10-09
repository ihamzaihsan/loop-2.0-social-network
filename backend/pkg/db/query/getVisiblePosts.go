package query

import "socialNetwork/pkg/models"

func GetVisiblePosts(user, limit, offset int) ([]models.PostResponse, int, error) {
	return VisiblePosts(user, 0, limit, offset, "")
}
