package dao

import "github.com/bugoutianzhen123/LetsTalk/domain"

type GroupDao interface {
	CreateGroup(group *domain.Group) error
	GetGroupById(groupId uint64) (domain.Group, error)
	GetGroupList() ([]domain.Group, error)
	AddGroupMember(member domain.GroupMember) error
	IsGroupMember(groupId, userId uint64) (bool, error)
	GetUserGroupIds(userId uint64) ([]uint64, error)
}

func (d *dao) CreateGroup(group *domain.Group) error {
	return d.db.Omit("Creator").Create(group).Error
}

func (d *dao) GetGroupById(groupId uint64) (domain.Group, error) {
	var group domain.Group
	err := d.db.Where("id = ?", groupId).First(&group).Error
	return group, err
}

func (d *dao) GetGroupList() ([]domain.Group, error) {
	var groups []domain.Group
	err := d.db.Order("id asc").Find(&groups).Error
	return groups, err
}

func (d *dao) AddGroupMember(member domain.GroupMember) error {
	return d.db.Omit("Group", "User", "Identity").Create(&member).Error
}

func (d *dao) IsGroupMember(groupId, userId uint64) (bool, error) {
	var count int64
	err := d.db.Model(&domain.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupId, userId).
		Count(&count).Error
	return count > 0, err
}

// GetUserGroupIds 返回用户已加入的群聊 ID 列表
func (d *dao) GetUserGroupIds(userId uint64) ([]uint64, error) {
	var ids []uint64
	err := d.db.Model(&domain.GroupMember{}).
		Where("user_id = ?", userId).
		Pluck("group_id", &ids).Error
	return ids, err
}
