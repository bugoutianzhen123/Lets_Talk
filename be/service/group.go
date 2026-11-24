package service

import (
	"errors"
	"strings"
	"time"

	"github.com/bugoutianzhen123/LetsTalk/domain"
)

type GroupService interface {
	CreateGroup(name string, creatorId uint64) (domain.Group, error)
	GetGroupList() ([]domain.Group, error)
	GetUserGroupIds(userId uint64) ([]uint64, error)
	JoinGroup(groupId, userId uint64) error
}

func (s *ser) CreateGroup(name string, creatorId uint64) (domain.Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Group{}, errors.New("群聊名称不能为空")
	}

	group := domain.Group{
		Name:      name,
		CreatorId: creatorId,
		Created:   time.Now(),
	}
	// 传指针，保证 GORM 写入的自增主键能回填到 group.Id
	if err := s.r.CreateGroup(&group); err != nil {
		return domain.Group{}, err
	}

	// 创建者自动成为群主
	if err := s.r.AddGroupMember(domain.GroupMember{
		GroupId:    group.Id,
		UserId:     creatorId,
		IdentityId: domain.GroupIdentityOwner,
		Created:    time.Now(),
	}); err != nil {
		return domain.Group{}, err
	}
	return group, nil
}

func (s *ser) GetGroupList() ([]domain.Group, error) {
	return s.r.GetGroupList()
}

func (s *ser) GetUserGroupIds(userId uint64) ([]uint64, error) {
	return s.r.GetUserGroupIds(userId)
}

// JoinGroup 主动加入一个已存在的群聊，加入者身份为一般群员
func (s *ser) JoinGroup(groupId, userId uint64) error {
	if _, err := s.r.GetGroupById(groupId); err != nil {
		return errors.New("该群聊不存在")
	}

	exists, err := s.r.IsGroupMember(groupId, userId)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("你已经在该群聊中了")
	}

	return s.r.AddGroupMember(domain.GroupMember{
		GroupId:    groupId,
		UserId:     userId,
		IdentityId: domain.GroupIdentityMember,
		Created:    time.Now(),
	})
}
