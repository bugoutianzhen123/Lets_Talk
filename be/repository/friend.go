package repository

import "github.com/bugoutianzhen123/LetsTalk/domain"

type Friend interface {
	CreateFriend(friend domain.Friend) error
	GetFriends(userId uint64) ([]domain.Friend, error)
	ExistsFriend(userId, friendId uint64) (bool, error)

	CreateFriendApplication(app domain.FriendApplication) error
	GetFriendApplicationById(id uint64) (domain.FriendApplication, error)
	GetPendingApplications(receiverId uint64) ([]domain.FriendApplication, error)
	UpdateFriendApplicationState(id uint64, stateId uint8) error
}

func (r *repo) CreateFriend(friend domain.Friend) error {
	return r.dao.CreateFriend(friend)
}

func (r *repo) GetFriends(userId uint64) ([]domain.Friend, error) {
	return r.dao.GetFriends(userId)
}

func (r *repo) ExistsFriend(userId, friendId uint64) (bool, error) {
	return r.dao.ExistsFriend(userId, friendId)
}

func (r *repo) CreateFriendApplication(app domain.FriendApplication) error {
	return r.dao.CreateFriendApplication(app)
}

func (r *repo) GetFriendApplicationById(id uint64) (domain.FriendApplication, error) {
	return r.dao.GetFriendApplicationById(id)
}

func (r *repo) GetPendingApplications(receiverId uint64) ([]domain.FriendApplication, error) {
	return r.dao.GetPendingApplications(receiverId)
}

func (r *repo) UpdateFriendApplicationState(id uint64, stateId uint8) error {
	return r.dao.UpdateFriendApplicationState(id, stateId)
}
