package repository

import "github.com/bugoutianzhen123/LetsTalk/domain"

// 私聊消息持久化接口
type FriendChat interface {
	SaveFriendMessage(message domain.FriendChatHistory) error
	GetFriendChatHistory(userId, friendId uint64) ([]domain.FriendChatHistory, error)
	GetLastFriendMessage(userId, friendId uint64) (domain.FriendChatHistory, error)
}

func (r *repo) SaveFriendMessage(message domain.FriendChatHistory) error {
	return r.dao.CreateFriendChatHistory(message)
}

func (r *repo) GetFriendChatHistory(userId, friendId uint64) ([]domain.FriendChatHistory, error) {
	return r.dao.GetFriendHistory(userId, friendId)
}

func (r *repo) GetLastFriendMessage(userId, friendId uint64) (domain.FriendChatHistory, error) {
	return r.dao.GetLastFriendMessage(userId, friendId)
}
