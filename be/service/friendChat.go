package service

import "github.com/bugoutianzhen123/LetsTalk/domain"

type FriendChatService interface {
	GetFriendChatHistory(userId, friendId uint64) ([]domain.FriendChatHistory, error)
	GetLastFriendMessage(userId, friendId uint64) (domain.FriendChatHistory, error)
	FriendManager() *FriendManager
	IsUserOnline(userId uint64) bool
}

func (s *ser) GetFriendChatHistory(userId, friendId uint64) ([]domain.FriendChatHistory, error) {
	return s.r.GetFriendChatHistory(userId, friendId)
}

func (s *ser) GetLastFriendMessage(userId, friendId uint64) (domain.FriendChatHistory, error) {
	return s.r.GetLastFriendMessage(userId, friendId)
}

func (s *ser) FriendManager() *FriendManager {
	return s.fm
}

// IsUserOnline 依据私聊常驻长连接判断用户是否在线
func (s *ser) IsUserOnline(userId uint64) bool {
	return s.fm.IsOnline(userId)
}
