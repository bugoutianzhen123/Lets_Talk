package dao

import "github.com/bugoutianzhen123/LetsTalk/domain"

type FriendDao interface {
	CreateFriend(friend domain.Friend) error
	GetFriends(userId uint64) ([]domain.Friend, error)
	ExistsFriend(userId, friendId uint64) (bool, error)

	CreateFriendApplication(app domain.FriendApplication) error
	GetFriendApplicationById(id uint64) (domain.FriendApplication, error)
	GetPendingApplications(receiverId uint64) ([]domain.FriendApplication, error)
	UpdateFriendApplicationState(id uint64, stateId uint8) error
}

func (d *dao) CreateFriend(friend domain.Friend) error {
	return d.db.Omit("User", "Friend").Create(&friend).Error
}

func (d *dao) GetFriends(userId uint64) ([]domain.Friend, error) {
	var friends []domain.Friend
	err := d.db.Where("user_id = ?", userId).
		Preload("Friend").
		Order("created desc").
		Find(&friends).Error
	return friends, err
}

func (d *dao) ExistsFriend(userId, friendId uint64) (bool, error) {
	var count int64
	err := d.db.Model(&domain.Friend{}).
		Where("user_id = ? AND friend_id = ?", userId, friendId).
		Count(&count).Error
	return count > 0, err
}

func (d *dao) CreateFriendApplication(app domain.FriendApplication) error {
	return d.db.Omit("Sender", "Receiver", "State").Create(&app).Error
}

func (d *dao) GetFriendApplicationById(id uint64) (domain.FriendApplication, error) {
	var app domain.FriendApplication
	err := d.db.Where("id = ?", id).First(&app).Error
	return app, err
}

func (d *dao) GetPendingApplications(receiverId uint64) ([]domain.FriendApplication, error) {
	var apps []domain.FriendApplication
	err := d.db.Where("receiver_id = ? AND state_id = ?", receiverId, domain.RequestStatePending).
		Preload("Sender").
		Order("created desc").
		Find(&apps).Error
	return apps, err
}

func (d *dao) UpdateFriendApplicationState(id uint64, stateId uint8) error {
	return d.db.Model(&domain.FriendApplication{}).
		Where("id = ?", id).
		Update("state_id", stateId).Error
}
