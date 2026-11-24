package dao

import "github.com/bugoutianzhen123/LetsTalk/domain"

type friendChatDao interface {
	CreateFriendChatHistory(message domain.FriendChatHistory) error
	GetFriendHistory(userId, friendId uint64) ([]domain.FriendChatHistory, error)
	GetLastFriendMessage(userId, friendId uint64) (domain.FriendChatHistory, error)
}

func (d *dao) CreateFriendChatHistory(message domain.FriendChatHistory) error {
	// Deleted 为零值时间，会导致 MySQL 严格模式写入失败，这里显式忽略
	return d.db.Omit("Sender", "Target", "Deleted").Create(&message).Error
}

func (d *dao) GetFriendHistory(userId, friendId uint64) ([]domain.FriendChatHistory, error) {
	var historys []domain.FriendChatHistory
	// 取双方之间的全部消息（我发给对方 + 对方发给我），按时间正序返回
	err := d.db.Where(
		"(sender_id = ? AND target_id = ?) OR (sender_id = ? AND target_id = ?)",
		userId, friendId, friendId, userId,
	).
		Preload("Sender").
		Order("created asc").
		Find(&historys).Error
	return historys, err
}

// GetLastFriendMessage 返回两人之间的最后一条消息（无消息时返回 gorm.ErrRecordNotFound）
func (d *dao) GetLastFriendMessage(userId, friendId uint64) (domain.FriendChatHistory, error) {
	var msg domain.FriendChatHistory
	err := d.db.Where(
		"(sender_id = ? AND target_id = ?) OR (sender_id = ? AND target_id = ?)",
		userId, friendId, friendId, userId,
	).
		Order("id desc").
		First(&msg).Error
	return msg, err
}
