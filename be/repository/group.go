package repository

import "github.com/bugoutianzhen123/LetsTalk/domain"

type Group interface {
	CreateGroup(group *domain.Group) error
	GetGroupById(groupId uint64) (domain.Group, error)
	GetGroupList() ([]domain.Group, error)
	AddGroupMember(member domain.GroupMember) error
	IsGroupMember(groupId, userId uint64) (bool, error)
	GetUserGroupIds(userId uint64) ([]uint64, error)
}

func (r *repo) CreateGroup(group *domain.Group) error {
	return r.dao.CreateGroup(group)
}

func (r *repo) GetGroupById(groupId uint64) (domain.Group, error) {
	return r.dao.GetGroupById(groupId)
}

func (r *repo) GetGroupList() ([]domain.Group, error) {
	return r.dao.GetGroupList()
}

func (r *repo) AddGroupMember(member domain.GroupMember) error {
	return r.dao.AddGroupMember(member)
}

func (r *repo) IsGroupMember(groupId, userId uint64) (bool, error) {
	return r.dao.IsGroupMember(groupId, userId)
}

func (r *repo) GetUserGroupIds(userId uint64) ([]uint64, error) {
	return r.dao.GetUserGroupIds(userId)
}
