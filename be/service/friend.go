package service

import (
	"errors"
	"time"

	"github.com/bugoutianzhen123/LetsTalk/domain"
)

type FriendService interface {
	ApplyFriend(senderId, receiverId uint64, message string) error
	GetPendingApplications(receiverId uint64) ([]domain.FriendApplication, error)
	AcceptFriendApplication(applicationId uint64) error
	GetFriends(userId uint64) ([]domain.Friend, error)
}

func (s *ser) ApplyFriend(senderId, receiverId uint64, message string) error {
	if senderId == receiverId {
		return errors.New("不能添加自己为好友")
	}
	if _, err := s.r.GetUserById(receiverId); err != nil {
		return errors.New("该用户不存在")
	}

	exists, err := s.r.ExistsFriend(senderId, receiverId)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("对方已经是你的好友了")
	}

	return s.r.CreateFriendApplication(domain.FriendApplication{
		Message:    message,
		SenderId:   senderId,
		ReceiverId: receiverId,
		StateId:    domain.RequestStatePending,
		Created:    time.Now(),
	})
}

func (s *ser) GetPendingApplications(receiverId uint64) ([]domain.FriendApplication, error) {
	return s.r.GetPendingApplications(receiverId)
}

func (s *ser) AcceptFriendApplication(applicationId uint64) error {
	app, err := s.r.GetFriendApplicationById(applicationId)
	if err != nil {
		return errors.New("好友申请不存在")
	}
	if app.StateId != domain.RequestStatePending {
		return errors.New("该申请已处理")
	}
	if err := s.r.UpdateFriendApplicationState(app.Id, domain.RequestStateAccepted); err != nil {
		return err
	}

	// 双向建立好友关系
	now := time.Now()
	if err := s.r.CreateFriend(domain.Friend{
		UserId:   app.SenderId,
		FriendId: app.ReceiverId,
		Created:  now,
		Updated:  now,
	}); err != nil {
		return err
	}
	return s.r.CreateFriend(domain.Friend{
		UserId:   app.ReceiverId,
		FriendId: app.SenderId,
		Created:  now,
		Updated:  now,
	})
}

func (s *ser) GetFriends(userId uint64) ([]domain.Friend, error) {
	return s.r.GetFriends(userId)
}
