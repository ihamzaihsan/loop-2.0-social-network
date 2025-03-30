package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetGroupMessages retrieves all messages for a group
func  GetGroupMessages(groupID int) ([]models.GroupMessage, error) {
    rows, err := db.DBInstance.DB.Query(`
        SELECT gcm.id, gcm.user_id as sender_id, gcm.content, gcm.created_at,
                u.id, u.first_name, u.last_name, u.avatar
        FROM group_chat_messages gcm
        JOIN users u ON gcm.user_id = u.id
        WHERE gcm.group_id = ?
        ORDER BY gcm.created_at ASC
    `, groupID)
    
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var messages []models.GroupMessage
    for rows.Next() {
        var msg models.GroupMessage
        var avatar sql.NullString
        
        err := rows.Scan(
            &msg.ID,
            &msg.SenderID,
            &msg.Content,
            &msg.CreatedAt,
            &msg.Sender.ID,
            &msg.Sender.FirstName,
            &msg.Sender.LastName,
            &avatar,
        )
        
        if err != nil {
            log.Printf("[ERROR] Failed to scan group message: %v", err)
            continue
        }
        
        if avatar.Valid {
            msg.Sender.Avatar = avatar.String
        }
        
        messages = append(messages, msg)
    }
    
    return messages, nil
}
