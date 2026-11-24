package dao

import "github.com/bugoutianzhen123/LetsTalk/domain"

type groupChatDao interface {
	CreateGroupChatHistory(message domain.GroupChatHistory) error
	GetGroupHistory(groupId uint64) ([]domain.GroupChatHistory, error)
	GetLastGroupMessage(groupId uint64) (domain.GroupChatHistory, error)
}

func (d *dao) CreateGroupChatHistory(message domain.GroupChatHistory) error {
	return d.db.Create(&message).Error
}

func (d *dao) GetGroupHistory(groupId uint64) ([]domain.GroupChatHistory, error) {
	var historys []domain.GroupChatHistory
	// 预加载发送者，便于前端展示用户名；按时间正序返回，方便直接渲染
	err := d.db.Where("group_id = ?", groupId).
		Preload("Sender").
		Order("created asc").
		Find(&historys).Error
	return historys, err
}

// GetLastGroupMessage 返回群聊的最后一条消息（无消息时返回 gorm.ErrRecordNotFound）
func (d *dao) GetLastGroupMessage(groupId uint64) (domain.GroupChatHistory, error) {
	var msg domain.GroupChatHistory
	err := d.db.Where("group_id = ?", groupId).Order("id desc").First(&msg).Error
	return msg, err
}
